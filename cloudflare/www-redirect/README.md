# Redirection du domaine www

Ce Worker répond en 301 pour `www.damiengoehrig.ca/*`, vers le même chemin et les mêmes paramètres sur `https://damiengoehrig.ca`. Il ne contacte aucune origine et ne traite aucun autre nom d’hôte. La route Cloudflare couvre HTTP et HTTPS ; la règle HTTPS globale de la zone peut ajouter une première redirection aux requêtes HTTP.

La connexion Wrangler existante autorise Workers et les routes, mais pas la modification des règles de redirection ou du DNS. Le Worker utilise donc la route www déjà proxifiée. Les autres routes Workers de la zone sont conservées.

Validation :

```sh
node --test cloudflare/www-redirect/redirect.test.mjs
npx wrangler deploy --dry-run --config cloudflare/www-redirect/wrangler.toml
```

Déploiement du Worker, indépendant du déploiement Hugo :

```sh
npx wrangler deploy --config cloudflare/www-redirect/wrangler.toml
```

Après déploiement, tester un article et ses paramètres, pas seulement l’accueil :

```sh
curl -I 'https://www.damiengoehrig.ca/fr/resume/?ref=cv'
```

Résultat attendu : 301 avec `Location: https://damiengoehrig.ca/fr/resume/?ref=cv`. En cas de retour arrière, retirer uniquement la route `www.damiengoehrig.ca/*` de ce Worker dans Cloudflare ; les autres routes et le domaine principal ne sont pas concernés.

Documentation : [Cloudflare — redirection avec un Worker](https://developers.cloudflare.com/workers/examples/redirect/), [routes Workers](https://developers.cloudflare.com/workers/configuration/routing/routes/).
