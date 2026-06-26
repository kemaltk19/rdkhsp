import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const srcDir = path.join(__dirname, 'src')

const replacements = {
  'label="Yeni Paket Ekle"': 'label="Yeni"',
  'label="Yeni Teklif Oluştur"': 'label="Yeni"',
  'label="Yeni Ürün/Hizmet"': 'label="Yeni"',
  'label="Yeni Cari Kart Ekle"': 'label="Yeni"',
  'label="Yeni Fatura Oluştur"': 'label="Yeni"',
  'label="Yeni Fatura Kes"': 'label="Yeni"',
  'label="Yeni Ödeme/Tahsilat"': 'label="Yeni"',
  'label="Yeni Fiş/Fatura"': 'label="Yeni"',
  'label="Yeni Çalışan"': 'label="Yeni"',
  
  'label="Faturayı Kesinleştir"': 'label="Kesinleştir"',
  'label="Teklifi Kaydet"': 'label="Kaydet"',
  'label="Faturayı Kaydet"': 'label="Kaydet"',
  'label="Cariyi Kaydet"': 'label="Kaydet"',
  'label="Profili Kaydet"': 'label="Kaydet"',
  'label="Fatura Ayarlarını Kaydet"': 'label="Kaydet"',
  'label="Teklif Ayarlarını Kaydet"': 'label="Kaydet"',
  'label="Cari Ayarlarını Kaydet"': 'label="Kaydet"',
  'label="Finans Ayarlarını Kaydet"': 'label="Kaydet"',
  
  'label="CSV Olarak Dışa Aktar"': 'label="Dışa Aktar"',
  'label="PDF İndir"': 'label="PDF"',
  'label="Faturaya Dönüştür"': 'label="Faturalandır"',
  
  'label="Tekliflere Dön"': 'label="Geri"',
  'label="Faturalara Dön"': 'label="Geri"',
  'label="Carilere Dön"': 'label="Geri"',
  'label="Ürünlere Dön"': 'label="Geri"',
  
  'label="Satır Ekle"': 'label="Ekle"',
  'label="Hareketi İşle"': 'label="Kaydet"',
  
  // also look for text between > and </Button> if any, though PrimeVue uses label prop.
}

function processDirectory(dir) {
  const files = fs.readdirSync(dir)
  
  for (const file of files) {
    const fullPath = path.join(dir, file)
    const stat = fs.statSync(fullPath)
    
    if (stat.isDirectory()) {
      processDirectory(fullPath)
    } else if (file.endsWith('.vue')) {
      processFile(fullPath)
    }
  }
}

function processFile(filePath) {
  let content = fs.readFileSync(filePath, 'utf8')
  let changed = false
  
  for (const [longLabel, shortLabel] of Object.entries(replacements)) {
    if (content.includes(longLabel)) {
      // replace all instances
      content = content.split(longLabel).join(shortLabel)
      changed = true
    }
  }
  
  if (changed) {
    fs.writeFileSync(filePath, content, 'utf8')
    console.log(`Shortened labels in: ${filePath}`)
  }
}

processDirectory(srcDir)
