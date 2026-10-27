import assert from "node:assert/strict";
import test from "node:test";
import worker from "./src/index.mjs";

test("redirects HTTP and HTTPS www requests to HTTPS apex", () => {
  for (const scheme of ["http", "https"]) {
    const response = worker.fetch(new Request(`${scheme}://www.damiengoehrig.ca/`));
    assert.equal(response.status, 301);
    assert.equal(response.headers.get("location"), "https://damiengoehrig.ca/");
  }
});

test("preserves encoded paths, repeated parameters and empty query values", () => {
  const suffix = "/fr/blog/un%20article/?ref=cv&tag=dbt&tag=ai&empty=";
  const response = worker.fetch(new Request("https://www.damiengoehrig.ca" + suffix));
  assert.equal(response.headers.get("location"), "https://damiengoehrig.ca" + suffix);
});

test("does not redirect unrelated hosts or loop on the apex", () => {
  for (const host of ["damiengoehrig.ca", "example.com", "www.damiengoehrig.ca.example.com"]) {
    assert.equal(worker.fetch(new Request(`https://${host}/`)).status, 404);
  }
});
