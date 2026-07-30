(function () {
  if (!window.Vue) return;

  window.MAX_UI_VUE_MAJOR = 3;

  if (typeof Vue.configureCompat === 'function') {
    Vue.configureCompat({
      MODE: 2,
      GLOBAL_MOUNT: 'suppress-warning',
      GLOBAL_EXTEND: 'suppress-warning',
      GLOBAL_PROTOTYPE: 'suppress-warning',
      GLOBAL_SET: 'suppress-warning',
      GLOBAL_DELETE: 'suppress-warning',
      CONFIG_OPTION_MERGE_STRATS: 'suppress-warning',
      INSTANCE_EVENT_EMITTER: 'suppress-warning',
      INSTANCE_EVENT_HOOKS: 'suppress-warning',
      OPTIONS_DATA_FN: 'suppress-warning',
      ATTR_FALSE_VALUE: 'suppress-warning',
      COMPILER_V_BIND_SYNC: 'suppress-warning',
      COMPILER_V_ON_NATIVE: 'suppress-warning',
      COMPILER_V_IF_V_FOR_PRECEDENCE: 'suppress-warning',
      COMPILER_NATIVE_TEMPLATE: 'suppress-warning',
      COMPILER_FILTERS: 'suppress-warning'
    });
  }

  Vue.config = Vue.config || {};
  Vue.config.productionTip = false;
})();
