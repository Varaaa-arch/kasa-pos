import {
  cartIncreaseButton,
  createAndTrackProduct,
  formatIDR,
  getProduct,
  getTransaction,
  productCardButton,
  test,
  expect,
  totalRow,
  waitForProductsLoaded,
} from "./fixtures";

test.describe("POS checkout — happy path", () => {
  test("cashier can complete a POS transaction", async ({
    page,
    api,
    testProducts,
  }) => {
    const product = await createAndTrackProduct(api, testProducts, {
      price: 15_000,
      stock: 10,
    });

    await page.goto("/");
    await waitForProductsLoaded(page);

    await page.getByLabel("Cari produk").fill(product.sku);
    await expect(productCardButton(page, product.name)).toBeVisible();

    await productCardButton(page, product.name).click();
    await cartIncreaseButton(page, product.name).click();

    await expect(page.getByText(`${formatIDR(product.price)} × 2`)).toBeVisible();
    await expect(totalRow(page)).toContainText(formatIDR(product.price * 2));

    const paidAmount = product.price * 2 + 8_000;
    await page.getByLabel("Bayar").fill(paidAmount.toString());

    await expect(page.getByText("Kembali").locator("..")).toContainText(
      formatIDR(8_000),
    );

    const checkoutResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/checkout") && response.request().method() === "POST",
    );

    await page.getByRole("button", { name: "BAYAR" }).click();
    const response = await checkoutResponse;
    expect(response.ok()).toBeTruthy();

    const successAlert = page
      .getByRole("alert")
      .filter({ hasText: "Transaksi berhasil" });
    await expect(successAlert).toBeVisible();
    await expect(successAlert).toContainText("ID transaksi:");
    await expect(successAlert).toContainText("No. invoice:");
    await expect(successAlert).toContainText(`Kembalian: ${formatIDR(8_000)}`);

    const transactionId = (
      await successAlert.getByText(/ID transaksi:/).textContent()
    )!
      .replace("ID transaksi:", "")
      .trim();

    await expect(page.getByText("Keranjang kosong")).toBeVisible();
    await expect(page.getByLabel("Bayar")).toHaveValue("");

    const updatedProduct = await getProduct(api, product.id);
    expect(updatedProduct.stock).toBe(product.stock - 2);

    const transaction = await getTransaction(api, transactionId);
    expect(transaction.Status).toBe("COMPLETED");
    expect(transaction.Total).toBe(product.price * 2);
    expect(transaction.PaidAmount).toBe(paidAmount);
    expect(transaction.Change).toBe(8_000);
    expect(transaction.Items).toHaveLength(1);
    expect(transaction.Items![0].ProductID).toBe(product.id);
    expect(transaction.Items![0].Quantity).toBe(2);
  });
});

test.describe("POS checkout — multi-item", () => {
  test("cashier can checkout multiple products with correct stock reduction", async ({
    page,
    api,
    testProducts,
  }) => {
    const suffix = Date.now().toString();
    const productA = await createAndTrackProduct(api, testProducts, {
      price: 12_000,
      stock: 20,
      suffix: `A-${suffix}`,
    });
    const productB = await createAndTrackProduct(api, testProducts, {
      price: 8_000,
      stock: 15,
      suffix: `B-${suffix}`,
    });

    await page.goto("/");
    await waitForProductsLoaded(page);

    await page.getByLabel("Cari produk").fill(productA.sku);
    await productCardButton(page, productA.name).click();

    await page.getByLabel("Cari produk").fill("");
    await page.getByLabel("Cari produk").fill(productB.sku);
    await productCardButton(page, productB.name).click();
    await cartIncreaseButton(page, productB.name).click();

    const expectedTotal = productA.price + productB.price * 2;
    await expect(totalRow(page)).toContainText(formatIDR(expectedTotal));

    const paidAmount = expectedTotal + 5_000;
    await page.getByLabel("Bayar").fill(paidAmount.toString());

    const checkoutResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/checkout") && response.request().method() === "POST",
    );
    await page.getByRole("button", { name: "BAYAR" }).click();
    expect((await checkoutResponse).ok()).toBeTruthy();

    await expect(
      page.getByRole("alert").filter({ hasText: "Transaksi berhasil" }),
    ).toBeVisible();

    const updatedA = await getProduct(api, productA.id);
    const updatedB = await getProduct(api, productB.id);
    expect(updatedA.stock).toBe(productA.stock - 1);
    expect(updatedB.stock).toBe(productB.stock - 2);
  });
});
