$files = @(
  "src\views\invoice\List.vue",
  "src\views\cari\List.vue",
  "src\views\expense\List.vue",
  "src\views\payment\List.vue",
  "src\views\quote\List.vue",
  "src\views\product\List.vue",
  "src\views\employee\List.vue",
  "src\views\report\List.vue",
  "src\views\settings\Index.vue",
  "src\views\billing\Billing.vue"
)

foreach ($f in $files) {
    $path = "c:\ai-folder\lite-Radikal-hesap\frontend\$f"
    if (Test-Path $path) {
        $content = Get-Content -Raw -Path $path
        # Remove the exact block
        $content = $content -replace '(?s)\s*<!-- Header -->\s*<div class="header-section mb-4">\s*<div>\s*<h1 class="page-title">.*?</div>\s*</div>', ''
        Set-Content -Path $path -Value $content
    }
}
