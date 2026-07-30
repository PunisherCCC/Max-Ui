(function () {
  if (!window.Vue) return;

  window.MAX_UI_VUE_MAJOR = 3;

  if (typeof Vue.configureCompat === 'function') {
    Vue.configureCompat({
      MODE: 2,
      ATTR_ENUMERATED_COERCION: 'suppress-warning',
      GLOBAL_MOUNT: 'suppress-warning',
      GLOBAL_MOUNT_CONTAINER: 'suppress-warning',
      GLOBAL_EXTEND: 'suppress-warning',
      GLOBAL_PROTOTYPE: 'suppress-warning',
      GLOBAL_SET: 'suppress-warning',
      GLOBAL_DELETE: 'suppress-warning',
      GLOBAL_OBSERVABLE: 'suppress-warning',
      GLOBAL_PRIVATE_UTIL: 'suppress-warning',
      CONFIG_OPTION_MERGE_STRATS: 'suppress-warning',
      CONFIG_PRODUCTION_TIP: 'suppress-warning',
      CONFIG_IGNORED_ELEMENTS: 'suppress-warning',
      CONFIG_KEY_CODES: 'suppress-warning',
      INSTANCE_EVENT_EMITTER: 'suppress-warning',
      INSTANCE_EVENT_HOOKS: 'suppress-warning',
      INSTANCE_SCOPED_SLOTS: 'suppress-warning',
      INSTANCE_ATTRS_CLASS_STYLE: 'suppress-warning',
      INSTANCE_CHILDREN: 'suppress-warning',
      INSTANCE_LISTENERS: 'suppress-warning',
      INSTANCE_SET: 'suppress-warning',
      INSTANCE_DELETE: 'suppress-warning',
      INSTANCE_DESTROY: 'suppress-warning',
      OPTIONS_DATA_FN: 'suppress-warning',
      OPTIONS_DATA_MERGE: 'suppress-warning',
      OPTIONS_BEFORE_DESTROY: 'suppress-warning',
      OPTIONS_DESTROYED: 'suppress-warning',
      ATTR_FALSE_VALUE: 'suppress-warning',
      COMPILER_V_BIND_SYNC: 'suppress-warning',
      COMPILER_V_ON_NATIVE: 'suppress-warning',
      COMPILER_V_IF_V_FOR_PRECEDENCE: 'suppress-warning',
      COMPILER_NATIVE_TEMPLATE: 'suppress-warning',
      COMPILER_INLINE_TEMPLATE: 'suppress-warning',
      COMPILER_FILTERS: 'suppress-warning',
      COMPILER_IS_ON_ELEMENT: 'suppress-warning',
      COMPILER_V_BIND_OBJECT_ORDER: 'suppress-warning',
      RENDER_FUNCTION: 'suppress-warning',
      TRANSITION_CLASSES: 'suppress-warning',
      TRANSITION_GROUP_ROOT: 'suppress-warning',
      TRANSITION_HOOK: 'suppress-warning',
      WATCH_ARRAY: 'suppress-warning',
      PRIVATE_APIS: 'suppress-warning'
    });
  }

  Vue.config = Vue.config || {};
  Vue.config.productionTip = false;
})();
