package service

import (
	"errors"
	"math"

	"github.com/mhsanaei/3x-ui/v2/database"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/logger"
	"github.com/mhsanaei/3x-ui/v2/xray"

	"gorm.io/gorm"
)

// Traffic Multiplier: a per-inbound policy that weights usage for accounting.
// It never throttles traffic; it only changes the bytes written to quota counters.
// The same multiplier is applied to upload and download, from the first byte.
//
// The multiplier is applied at exactly two choke points, addClientTraffic (the 10s
// collection tick) and foldClientTraffic (the RADIUS teardown paths), and nowhere
// else. Applying it inside a protocol collector as well would double-multiply.

// MaxTrafficMultiplier bounds the weight. Far above any plausible billing policy,
// and far below the point where one tick's bytes times the multiplier could
// overflow int64 (a 10s tick at 10Gbps is ~1.2e10 bytes; 1e13 leaves six orders of
// headroom under int64's 9.2e18).
const MaxTrafficMultiplier = 1000

// validMultiplier reports whether m is a usable weight.
//
// NaN is the dangerous case and the reason this is a named check rather than an
// inline comparison: `NaN <= 1` is FALSE, so a bare `m <= 1` guard lets NaN
// straight through, and int64(NaN) is MinInt64. That drives a client's counter to
// roughly -4.6e18, and the enforcement predicate (`up + down >= total`) is then
// false forever, so the account can never be quota-disabled again and its counter
// is unrecoverable. Inf and any value large enough to overflow int64 do the same.
func validMultiplier(m float64) bool {
	return !math.IsNaN(m) && !math.IsInf(m, 0) && m > 1 && m <= MaxTrafficMultiplier
}

// multiplyDelta weights one raw byte delta according to inb's policy. The
// currentUpDown parameter is retained for compatibility with older tests/callers,
// but no threshold is used anymore: when enabled, every byte is billed at the
// inbound's multiplier from byte zero.
// Returns the deltas unchanged when the policy is off, so callers can call it
// unconditionally.
func multiplyDelta(inb *model.Inbound, currentUpDown, deltaUp, deltaDown int64) (int64, int64) {
	// Saves are validated (validateInboundConfig), but this is the last line of
	// defence and it must hold for a row that arrived some other way: an imported
	// DB, a hand-edited SQLite file, a future caller. Billing raw is always safe;
	// billing NaN is unrecoverable.
	if inb == nil || !inb.TrafficMultiplierEnable || !validMultiplier(inb.TrafficMultiplier) {
		return deltaUp, deltaDown
	}
	raw := deltaUp + deltaDown
	if raw <= 0 {
		return deltaUp, deltaDown
	}
	_ = currentUpDown
	billed := int64(math.Round(float64(raw) * inb.TrafficMultiplier))
	// Apportion the billed total back across the two directions so their ratio stays
	// honest in reporting. Only the sum is ever enforced, so the split is cosmetic;
	// the remainder goes to Down to keep up+down == billed exactly.
	up := int64(math.Round(float64(billed) * (float64(deltaUp) / float64(raw))))
	return up, billed - up
}

// multiplierColumns is the minimal column set multiplyDelta needs. Selecting it
// explicitly keeps the 10s tick from loading every inbound's Settings JSON blob.
var multiplierColumns = []string{"id", "traffic_multiplier_enable", "traffic_multiplier_after", "traffic_multiplier"}

// loadMultiplierInbounds batch-loads the inbounds owning the given client rows,
// keyed by inbound id. One query for the whole tick rather than one per client.
func loadMultiplierInbounds(tx *gorm.DB, cts []*xray.ClientTraffic) (map[int]*model.Inbound, error) {
	ids := make([]int, 0, len(cts))
	seen := make(map[int]struct{}, len(cts))
	for _, ct := range cts {
		if ct.InboundId == 0 {
			continue
		}
		if _, dup := seen[ct.InboundId]; dup {
			continue
		}
		seen[ct.InboundId] = struct{}{}
		ids = append(ids, ct.InboundId)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var inbounds []*model.Inbound
	err := tx.Model(model.Inbound{}).Select(multiplierColumns).Where("id IN (?)", ids).Find(&inbounds).Error
	if err != nil {
		return nil, err
	}
	byID := make(map[int]*model.Inbound, len(inbounds))
	for _, ib := range inbounds {
		byID[ib.Id] = ib
	}
	return byID, nil
}

// foldClientTraffic adds a torn-down session's final bytes to a client's counters,
// applying the owning inbound's traffic multiplier.
//
// The RADIUS teardown paths (acct-stop, rbridge reconcile, user-limit evict) flush
// bytes outside the 10s collection tick, so they need the multiplier applied here
// or that traffic bills at 1:1 forever. The transaction keeps the final raw
// lifetime counter and billed quota counters updated together.
func foldClientTraffic(email string, up, down int64) {
	if email == "" || (up <= 0 && down <= 0) {
		return
	}
	db := database.GetDB()
	if db == nil {
		return
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		var ct xray.ClientTraffic
		err := tx.Model(xray.ClientTraffic{}).Where("email = ?", email).First(&ct).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // no row to bill, same as the plain UPDATE matching nothing
		}
		if err != nil {
			return err
		}
		billedUp, billedDown := up, down
		if ct.InboundId != 0 {
			var inb model.Inbound
			err := tx.Model(model.Inbound{}).Select(multiplierColumns).
				Where("id = ?", ct.InboundId).First(&inb).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err == nil {
				billedUp, billedDown = multiplyDelta(&inb, ct.Up+ct.Down, up, down)
			}
		}
		// all_time takes the RAW delta while up/down take the billed one, matching
		// addClientTraffic. It is written here rather than left alone because the
		// startup backfill (MigrationRequirements) sets all_time = up+down for any
		// row still at 0, which for a client whose traffic only ever arrived through
		// this path would seed the lifetime record with MULTIPLIED bytes.
		return tx.Exec(
			"UPDATE client_traffics SET up = up + ?, down = down + ?, all_time = COALESCE(all_time, 0) + ? WHERE email = ?",
			billedUp, billedDown, up+down, email).Error
	})
	if err != nil {
		logger.Warning("fold client traffic for ", email, ": ", err)
	}
}
