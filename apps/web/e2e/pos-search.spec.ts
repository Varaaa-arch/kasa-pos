import {
  createAndTrackProduct,
  productCardButton,
  test,
  expect,
  waitForProductsLoaded,
} from "./fixtures";

test.describe("POS product search", () => {
  test("filters products by name and SKU", async ({ page, api, testProducts }) => {
    const suffix = Date.now().toString();
    const product = await createAndTrackProduct(api, testProducts, {
      price: 10_000,
      stock: 5,
      suffix,
    });

    await page.goto("/");
    await waitForProductsLoaded(page);

    const searchInput = page.getByLabel("Cari produk");

    await searchInput.fill(product.name);
    await expect(productCardButton(page, product.name)).toBeVisible();

    await searchInput.fill(product.sku);
    await expect(productCardButton(page, product.name)).toBeVisible();

    await searchInput.fill("E2E-NONEXISTENT-PRODUCT-XYZ");
    await expect(page.getByText("Produk tidak ditemukan")).toBeVisible();
    await expect(productCardButton(page, product.name)).toBeHidden();
  });
});
