# Maintenance SEO

Le site conserve son thème Blowfish v2.93.0. Les correctifs du 27 octobre 2026 sont des surcharges locales du thème ; la refonte précédente reste dans le stash `avant-audit-seo-2026-10-27`.

## Vérifier un changement

```sh
make build-hugo-only
make check-redirect
```

`make build`, utilisé par Cloudflare Pages, valide aussi les pages générées après le build. `scripts/check-seo.py` utilise uniquement la bibliothèque standard Python et contrôle les canonicals, les traductions réciproques, les types et propriétés des schémas, les images sociales, les titres de profil, les H1 et les sitemaps, ainsi que les redirections et la présence des articles dans les listes (comparaison avec les flux RSS). Les mêmes contrôles sont exécutés en CI.

Pour écrire un inventaire des archives sans décider de leur désindexation :

```sh
python3 scripts/check-seo.py public --inventory /tmp/dad-driven-development-taxonomies.json
```

## Métadonnées et traductions

- `layouts/partials/seo/metadata.html` fournit titre, description, canonical et image à la fois au head HTML, aux aperçus sociaux et au JSON-LD. Cette donnée dépend de la page et ne doit pas être mise en cache par site.
- `seoTitle` personnalise le titre de recherche sans changer le nom visible sur l’accueil ou le CV. `description` est prioritaire sur la longue synthèse du CV.
- Les articles traduits dont les slugs diffèrent doivent partager une `translationKey`. Les hreflang doivent être réciproques et rester sur la même page de pagination.
- L’accueil affiche une sélection non paginée. La pagination du blog possède ses propres canonicals. Les anciennes routes `/page/2/` et `/fr/page/2/` redirigent vers les archives correspondantes via `static/_redirects`.
- Les billets portent `BlogPosting`, le CV `ProfilePage`, les archives `CollectionPage` et l’accueil `WebSite`. Ne pas décrire une archive comme un article.
- L’image propre à chaque article reste prioritaire. Le portrait paysage existant sert de repli aux pages sans couverture.

Lors d’une mise à jour de Blowfish, comparer notamment les surcharges `head.html`, `header/basic.html`, `home/profile.html` et `recent-articles/cardview.html` avec le thème, puis refaire un parcours clavier et les contrôles visuels.

## Navigation et mesure

Le menu mobile utilise de vrais boutons, expose son état, gère Échap et contient le focus pendant son ouverture. Conserver ces comportements lors d’une modification du JavaScript.

Les liens explicites CV/contact de l’accueil envoient l’événement GA4 `profile_action` avec `action` (`resume` ou `contact`) et `page_language`. L’envoi utilise l’intégration GA4 différée existante ; sa réception n’a pas été vérifiée dans une propriété Analytics.

Cloudflare Pages injecte le beacon Web Analytics configuré dans le projet. Ne pas réintroduire une deuxième balise manuelle dans le thème ou la CI. Sur un aperçu `pages.dev`, cet envoi peut échouer en CORS car l’origine diffère du domaine de production : ce n’est pas une preuve d’échec sur le domaine canonique.

## Domaine www

Le Worker dédié et ses tests sont dans [cloudflare/www-redirect](cloudflare/www-redirect/README.md). La redirection est indépendante de Hugo et conserve le chemin et la query string. Le fichier Pages `_redirects` ne prend pas en charge les redirections par nom d’hôte.

## Archives et données Search Console

La revue GSC du 27 octobre 2026 est documentée dans `damiengoehrig.ca-audit/gsc/REPORT.md`. Huit archives strictement dupliquées et leurs flux RSS redirigent vers les destinations retenues ; sept suivent la canonicale inspectée dans Google, la huitième son équivalent français. Le plan détaillé et les ensembles d’articles sont conservés dans `damiengoehrig.ca-audit/implementation/archive-consolidation.json`. Les archives ayant reçu des clics sont conservées.

Les rubriques dbt, IA, ingénierie data, open source, qualité et gouvernance ont une introduction EN/FR. `seoHub: true` les inclut dans le sitemap malgré l’exclusion par défaut des taxonomies du thème. La gouvernance utilise une catégorie EN et un tag FR, avec une `translationKey` commune ; conserver cette association.

La pagination doit être initialisée avec `.Paginate (.Pages.GroupByDate "2006")`, comme les listes Blowfish. Appeler `.Paginator` dans le head avant cette initialisation peut vider `.PageGroups`. Le contrôle RSS/HTML protège le contenu réel des archives, en plus des métadonnées.

Avant toute nouvelle fusion, vérifier les ensembles d’articles, les clics, impressions et liens entrants. Les sources redirigées doivent sortir des sitemaps et les hreflang doivent viser les destinations finales.

Après publication, vérifier les URLs publiques et leurs sitemaps, puis suivre les impressions Data/dbt et les visites vers CV/contact pendant quatre à huit semaines. GSC confirme la présence des 26 articles dans les exemples indexés. Les Core Web Vitals terrain restent indisponibles faute de volume ; la prise en compte des nouveaux correctifs reste à suivre après publication.
