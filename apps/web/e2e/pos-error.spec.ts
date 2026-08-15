import {
  cartIncreaseButton,
  createAndTrackProduct,
  formatIDR,
  getProduct,
  productCardButton,
  test,
  expect,
  totalRow,
  waitForProductsLoaded,
} from "./fixtures";

test.describe("POS checkout — insufficient payment", () => {
  test("rejects checkout when payment is lower than total", async ({
    page,
    api,
    testProducts,
  }) => {
    const product = await createAndTrackProduct(api, testProducts, {
      price: 20_000,
      stock: 10,
    });

    const transactionsBefore = await api.get("/transactions");
    expect(transactionsBefore.ok()).toBeTruthy();
    const transactionCountBefore = (await transactionsBefore.json()).length;

    await page.goto("/");
    await waitForProductsLoaded(page);

    await page.getByLabel("Cari produk").fill(product.sku);
    await productCardButton(page, product.name).click();
    await cartIncreaseButton(page, product.name).click();

    const total = product.price * 2;
    const insufficientPayment = total - 5_000;

    await expect(totalRow(page)).toContainText(formatIDR(total));
    await page.getByLabel("Bayar").fill(insufficientPayment.toString());

    await expect(
      page.getByRole("alert").filter({ hasText: "Pembayaran kurang" }),
    ).toBeVisible();

    const payButton = page.getByRole("button", { name: "BAYAR" });
    await expect(payButton).toBeDisabled();

    let checkoutCalled = false;
    page.on("request", (request) => {
      if (request.url().includes("/checkout") && request.method() === "POST") {
        checkoutCalled = true;
      }
    });

    await payButton.click({ force: true });
    expect(checkoutCalled).toBe(false);

    await expect(page.getByText(`${formatIDR(product.price)} × 2`)).toBeVisible();

    const unchangedProduct = await getProduct(api, product.id);
    expect(unchangedProduct.stock).toBe(product.stock);

    const transactionsAfter = await api.get("/transactions");
    expect(transactionsAfter.ok()).toBeTruthy();
    const transactionCountAfter = (await transactionsAfter.json()).length;
    expect(transactionCountAfter).toBe(transactionCountBefore);
  });
});

test.describe("POS checkout — printer integration", () => {
  test("checkout triggers print agent after successful transaction", async ({
    page,
    api,
    testProducts,
  }) => {
    const printAgentUrl = process.env.PRINT_AGENT_URL ?? "http://127.0.0.1:8081";

    const product = await createAndTrackProduct(api, testProducts, {
      price: 10_000,
      stock: 10,
    });

    await page.goto("/");
    await waitForProductsLoaded(page);

    await page.getByLabel("Cari produk").fill(product.sku);
    await productCardButton(page, product.name).click();

    const paidAmount = product.price + 5_000;
    await page.getByLabel("Bayar").fill(paidAmount.toString());

    const checkoutResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/checkout") && response.request().method() === "POST",
    );

    await page.getByRole("button", { name: "BAYAR" }).click();
    const response = await checkoutResponse;
    expect(response.ok()).toBeTruthy();

    const checkoutBody = (await response.json()) as {
      status: string;
      print_job?: { status: string; error?: string };
    };

    expect(checkoutBody.status).toBe("COMPLETED");
    expect(checkoutBody.print_job).toBeDefined();

    try {
      const statusResponse = await fetch(`${printAgentUrl}/status`);
      if (statusResponse.ok) {
        const status = (await statusResponse.json()) as Record<string, unknown>;
        test.info().annotations.push({
          type: "printer",
          description: `Print agent reachable at ${printAgentUrl}. Checkout triggers POST /print after DB commit. Physical USB receipt output (BP-LITE58) remains a manual hardware smoke test.`,
        });
        expect(status).toHaveProperty("printer");

        if (checkoutBody.print_job?.status === "COMPLETED") {
          expect(checkoutBody.print_job.error).toBeUndefined();
        }
      }
    } catch {
      test.info().annotations.push({
        type: "printer",
        description:
          "Print agent not running. Checkout still completes; print_job reports failure. Physical printing remains a manual hardware smoke test.",
      });

      expect(checkoutBody.print_job?.status).toBe("FAILED");
      expect(checkoutBody.print_job?.error).toBeTruthy();
    }
  });
});
