(function() {
  var cfg = window.GA_CONFIG || {};
  var enabled = !!(cfg.measurementId || cfg.adsId);

  if (enabled) {
    window.dataLayer = window.dataLayer || [];
    window.gtag = function() { dataLayer.push(arguments); };

    gtag('js', new Date());

    if (cfg.measurementId) {
      gtag('config', cfg.measurementId, {
        analytics_storage: 'denied',
        ad_storage: 'denied',
        send_page_view: false
      });
    }

    if (cfg.adsId) {
      gtag('config', cfg.adsId);
    }
  }

  window.Analytics = {
    trackPageView: function(path) {
      if (!enabled) return;
      gtag('event', 'page_view', {
        page_path: path,
        page_location: window.location.origin + path,
        page_title: document.title
      });
    },

    trackEvent: function(name, params) {
      if (!enabled) return;
      gtag('event', name, params || {});
      if (name === 'extraction_completed' || name === 'download_completed') {
        this.trackConversion();
      }
    },

    trackConversion: function(label, value, currency) {
      if (!cfg.adsId || !cfg.adsLabel) return;
      gtag('event', 'conversion', {
        send_to: cfg.adsId + '/' + (label || cfg.adsLabel),
        value: value || 0,
        currency: currency || 'USD'
      });
    }
  };

  if (cfg.measurementId) {
    Analytics.trackPageView(window.location.pathname + window.location.search);

    var origPushState = history.pushState;
    history.pushState = function() {
      origPushState.apply(this, arguments);
      Analytics.trackPageView(window.location.pathname + window.location.search);
    };
    window.addEventListener('popstate', function() {
      Analytics.trackPageView(window.location.pathname + window.location.search);
    });
  }
})();
