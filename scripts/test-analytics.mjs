import fs from 'node:fs';
import vm from 'node:vm';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
const root=fileURLToPath(new URL('../',import.meta.url));
const source=fs.readFileSync(root+'/assets/js/analytics.js','utf8');
function tracker({host='damiengoehrig.ca',article=false}={}) {
  let now=0, interval, bounds={top:0,height:1000};const events=[],listeners=new Map(),windowListeners=new Map();
  const context={content_group:article?'article':'portfolio',content_id:article?'souffle':'home',page_locale:'fr'};
  const window={siteAnalyticsContext:context,siteAnalyticsProjects:[{name:'Soufflé',url:'https://usesouffle.app/fr/'},{name:'dbt-guard',url:'https://github.com/damione1/dbt-guard'}],gtag:(...args)=>events.push(JSON.parse(JSON.stringify(args))),addEventListener:(name,fn)=>windowListeners.set(name,fn)};
  const document={documentElement:{lang:'fr'},visibilityState:'visible',getElementById:()=>null,querySelector:()=>({getBoundingClientRect:()=>bounds}),addEventListener:(name,fn,options)=>listeners.set(name,{fn,options})};
  vm.runInNewContext(source,{window,document,location:new URL('https://'+host+'/fr/blog/test/'),URL,sessionStorage:{getItem:()=>null},performance:{now:()=>now},innerHeight:500,setInterval(fn){interval=fn;return 1;},clearInterval(){}});
  return {window,document,events,listeners,windowListeners,track:window.siteAnalytics.track,clock:value=>now=value,progress:value=>bounds=value,check:()=>interval()};
}
function click(state,href,{profileAction,button=0,type='click'}={}) {
  const link={href,dataset:{profileAction},closest:()=>null};state.listeners.get(type).fn({type,button,target:{closest:()=>link}});
}
const s=tracker();assert.equal(s.listeners.get('click').options.capture,true);
click(s,'https://usesouffle.app/docs/getting-started/');
click(s,'https://github.com/damione1/dbt-guard/tree/master');
click(s,'https://github.com/damione1/dbt-guard-other');
click(s,'https://www.linkedin.com/in/damiengoehrig/');
click(s,'https://www.linkedin.com/sharing/share-offsite/?url=example');
click(s,'https://damiengoehrig.ca/fr/resume/',{profileAction:'resume'});
assert.deepEqual(s.events.map(e=>e[1]),['project_visit','project_visit','linkedin_visit','profile_action']);
assert.equal(s.events[0][2].project_name,'Soufflé');assert.equal(s.events[3][2].page_language,'fr');
s.track('project_visit',{clipboard:'secret',search_term:'secret'});assert(!JSON.stringify(s.events).includes('secret'));
const local=tracker({host:'localhost'});local.track('article_read');assert.equal(local.events.length,0);
const a=tracker({article:true});a.document.visibilityState='hidden';a.listeners.get('visibilitychange').fn();a.clock(60000);a.check();assert.equal(a.events.length,0);
a.document.visibilityState='visible';a.listeners.get('visibilitychange').fn();a.clock(90000);a.check();assert.equal(a.events.length,0);
a.progress({top:-300,height:1000});a.windowListeners.get('scroll')();a.check();assert.equal(a.events.length,1);assert.equal(a.events[0][1],'article_read');assert.equal(a.events[0][2].value,30);
function contextOf(path){const html=fs.readFileSync(root+'/public/'+path,'utf8');const script=[...html.matchAll(/<script\b[^>]*>([\s\S]*?)<\/script>/g)].map(m=>m[1]).find(s=>s.includes('window.siteAnalyticsContext='));assert(script);const window={};vm.runInNewContext(script,{window,location:{hostname:'localhost'}});return window;}
const en=contextOf('blog/souffle-local-transcription-macos/index.html'),fr=contextOf('fr/blog/souffle-transcription-locale-macos/index.html');
assert.equal(en.siteAnalyticsContext.content_group,'article');assert.equal(fr.siteAnalyticsContext.page_locale,'fr');assert.equal(en.siteAnalyticsContext.content_id,fr.siteAnalyticsContext.content_id);assert(en.siteAnalyticsProjects.length >= 8);
console.log('PASS: project/profile matching, modal capture, legacy event preservation, no free text, local gating, foreground reading/progress, bilingual article identity and complete project catalogue.');
