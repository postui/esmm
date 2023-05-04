import {
  assertEquals,
  assertStringIncludes,
} from "https://deno.land/std@0.180.0/testing/asserts.ts";

import { build, esm } from "http://localhost:8080/build";

Deno.test("build api", async (t) => {
  let url = "";
  let bundleUrl = "";
  await t.step("build", async () => {
    const ret = await fetch("http://localhost:8080/build", {
      method: "POST",
      headers: { "Content-Type": "application/javascript" },
      body: `export default "Hello world!";`,
    }).then((r) => r.json());
    if (ret.error) {
      throw new Error(`<${ret.error.status}> ${ret.error.message}`);
    }
    url = ret.url;
    bundleUrl = ret.bundleUrl;
    assertStringIncludes(url, "/~");
    assertStringIncludes(bundleUrl, "?bundle");
  });

  await t.step("import published module", async () => {
    const { default: message } = await import(url);
    assertEquals(message, "Hello world!");
  });
});

Deno.test("build api (json)", async (t) => {
  let url = "";
  let bundleUrl = "";
  await t.step("build", async () => {
    const ret = await fetch("http://localhost:8080/build", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        code: `/* @jsx h */
          import { h } from "npm:preact@10.13.2";
          import { renderToString } from "npm:preact-render-to-string@6.0.2";
          export default () => renderToString(<h1>Hello world!</h1>);
        `,
      }),
    }).then((r) => r.json());
    if (ret.error) {
      throw new Error(`<${ret.error.status}> ${ret.error.message}`);
    }
    url = ret.url;
    bundleUrl = ret.bundleUrl;
    assertStringIncludes(url, "/~");
    assertStringIncludes(bundleUrl, "?bundle");
  });

  await t.step("import published module", async () => {
    const { default: render1 } = await import(url);
    const { default: render2 } = await import(bundleUrl);
    assertEquals(render1(), "<h1>Hello world!</h1>");
    assertEquals(render2(), "<h1>Hello world!</h1>");
  });
});

Deno.test("build api (use sdk)", async (t) => {
  await t.step("use `build` function", async () => {
    const ret = await build(`export default "Hello world!";`);
    const { default: message } = await import(ret.url);
    assertEquals(message, "Hello world!");
  });

  await t.step("use `esm` template function", async () => {
    const message = "Hello world!";
    const ret = await esm`export default ${JSON.stringify(message)};`;
    const mod = await import(ret.url);
    assertEquals(mod.default, message);
  });
});
