// SPDX-License-Identifier: GPL-3.0-or-later
(() => {
  if (window.siteAnalytics) return;
  const context = window.siteAnalyticsContext;
  const events = new Set(['article_read','project_visit','linkedin_visit','github_visit','profile_action',
    'search_open','view_search_results','search_no_results','search_result_click','web_vital']);
  const allowed = new Set(['project_name','link_placement','content_type','content_id','action','page_language',
    'result_count','result_position','destination_path','metric_name','metric_rating','metric_id','metric_value','value']);
  const vitalsSource = document.currentScript?.dataset.webVitals;
  function track(name, parameters = {}) {
    if (!events.has(name) || location.hostname !== 'damiengoehrig.ca' || typeof window.gtag !== 'function') return;
    const values = {...context};
    for (const [key,value] of Object.entries(parameters)) {
      if (!allowed.has(key)) continue;
      if (typeof value === 'string' && value.length > 0 && value.length <= 100) values[key] = value;
      else if (typeof value === 'number' && Number.isFinite(value) && (value >= 0 || (name === 'web_vital' && key === 'value'))) values[key] = value;
    }
    try { if (sessionStorage.getItem('site-analytics-debug') === '1') values.debug_mode = true; } catch {}
    window.gtag('event',name,values);
  }
  window.siteAnalytics = {track};
  const catalogue = window.siteAnalyticsProjects.map(project => ({name:project.name,url:new URL(project.url,location.href)}));
  const placement = element => element.closest('#search-results') ? 'search'
    : element.closest('header') || element.closest('nav') ? 'navigation'
    : element.closest('footer') ? 'footer' : 'content';
  function click(event) {
    if ((event.type === 'auxclick' && event.button !== 1) || (event.type === 'click' && event.button > 0)) return;
    const link = event.target.closest?.('a[href]');
    if (!link) return;
    let url;
    try { url = new URL(link.href,location.href); } catch { return; }
    const parameters = {link_placement:placement(link)};
    if (link.dataset.profileAction) {
      // Retain the existing event and parameters for historical reports.
      track('profile_action',{action:link.dataset.profileAction,page_language:document.documentElement.lang,...parameters});
    }
    const project = catalogue.find(project => {
      if (url.origin !== project.url.origin) return false;
      if (url.hostname !== 'github.com') return true;
      const prefix = project.url.pathname.replace(/\/$/,'');
      return url.pathname === prefix || url.pathname.startsWith(prefix + '/');
    });
    if (project) track('project_visit',{...parameters,project_name:project.name,content_type:'project'});
    else if (/^(?:www\.)?linkedin\.com$/.test(url.hostname) && /^\/in\/damiengoehrig\/?$/.test(url.pathname)) track('linkedin_visit',parameters);
    else if (url.hostname === 'github.com' && /^\/damione1\/?$/.test(url.pathname)) track('github_visit',parameters);
    if (link.closest('#search-results')) {
      const links = [...document.querySelectorAll('#search-results a[href]')];
      track('search_result_click',{destination_path:url.pathname,result_position:links.indexOf(link) + 1});
    }
  }
  // The installed search modal stops bubbling; observe intents in capture phase.
  document.addEventListener('click',click,{capture:true});
  document.addEventListener('auxclick',click,{capture:true});

  if (context.content_group === 'article') {
    const article = document.querySelector('.article-content');
    let activeMs = 0, started = document.visibilityState === 'visible' ? performance.now() : null, reported = false;
    const check = () => {
      const now = performance.now();
      if (started !== null) activeMs += now - started;
      started = document.visibilityState === 'visible' ? now : null;
      if (!article || reported || activeMs < 30000) return;
      const box = article.getBoundingClientRect();
      if ((innerHeight - box.top) / Math.max(1,box.height) < 0.75) return;
      reported = true; clearInterval(timer); track('article_read',{value:Math.floor(activeMs / 1000)});
    };
    const timer = setInterval(check,1000);
    document.addEventListener('visibilitychange',check);
    window.addEventListener('scroll',check,{passive:true});
    window.addEventListener('pagehide',check);
  }

  const wrapper = document.getElementById('search-wrapper'), input = document.getElementById('search-query');
  if (wrapper && input) {
    let open = wrapper.style.visibility === 'visible', measurement, revision = 0, reportedQuery = null;
    const schedule = () => {
      clearTimeout(measurement);
      if (!input.value.trim()) reportedQuery = null;
      const current = ++revision;
      measurement = setTimeout(() => {
        // The installed Blowfish search engine sets indexed after its JSON index loads.
        if (current !== revision || wrapper.style.visibility !== 'visible' || !input.value.trim() || !window.indexed || input.value.trim() === reportedQuery) return;
        reportedQuery = input.value.trim();
        const count = document.querySelectorAll('#search-results a[href]').length;
        track('view_search_results',{result_count:count,value:count});
        if (!count) track('search_no_results');
      },600);
    };
    new MutationObserver(() => {
      const visible = wrapper.style.visibility === 'visible';
      if (visible && !open) track('search_open');
      if (!visible) {++revision;clearTimeout(measurement);reportedQuery=null;}
      open = visible;
    }).observe(wrapper,{attributes:true,attributeFilter:['style']});
    input.addEventListener('keyup',schedule);
  }
  if (vitalsSource && location.hostname === 'damiengoehrig.ca') {
    import(vitalsSource).then(({onLCP,onINP,onCLS}) => {
      const report = metric => track('web_vital',{metric_name:metric.name,metric_rating:metric.rating,
        metric_id:metric.id,metric_value:metric.value,value:metric.delta});
      onLCP(report);onINP(report);onCLS(report);
    }).catch(() => {});
  }
})();
