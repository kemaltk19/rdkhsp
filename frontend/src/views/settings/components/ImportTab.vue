<script setup lang="ts">
import { ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import FileUpload from 'primevue/fileupload'
import { importCarisApi, importProductsApi, downloadSampleCariApi, downloadSampleProductApi } from '@/api/settings'

const toast = useToast()

const downloadingCari = ref(false)
const downloadingProduct = ref(false)

const handleDownloadSampleCari = async () => {
  downloadingCari.value = true
  try {
    const res = await downloadSampleCariApi()
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'Cari_Iceri_Aktarim_Sablonu.xlsx')
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    toast.add({ severity: 'success', summary: 'İndirildi', detail: 'Örnek Excel dosyası indirildi.', life: 10000 })
  } catch (error) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Örnek dosya indirilemedi.', life: 10000 })
  } finally {
    downloadingCari.value = false
  }
}

const handleDownloadSampleProduct = async () => {
  downloadingProduct.value = true
  try {
    const res = await downloadSampleProductApi()
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'Stok_Karti_Iceri_Aktarim_Sablonu.xlsx')
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    toast.add({ severity: 'success', summary: 'İndirildi', detail: 'Örnek Excel dosyası indirildi.', life: 10000 })
  } catch (error) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Örnek dosya indirilemedi.', life: 10000 })
  } finally {
    downloadingProduct.value = false
  }
}

const customUploadCari = async (event: any) => {
  const file = event.files[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  try {
    const res = await importCarisApi(formData)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: res.data.message || 'Cari kartları içe aktarıldı.', life: 10000 })
    event.options.clear()
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: error.response?.data?.error || 'Dosya yüklenirken bir hata oluştu.', life: 10000 })
    event.options.clear()
  }
}

const customUploadProduct = async (event: any) => {
  const file = event.files[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  try {
    const res = await importProductsApi(formData)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: res.data.message || 'Stok kartları içe aktarıldı.', life: 10000 })
    event.options.clear()
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: error.response?.data?.error || 'Dosya yüklenirken bir hata oluştu.', life: 10000 })
    event.options.clear()
  }
}
</script>

<template>
  <div class="space-y-6">
    <Card>
      <template #title>
        <div class="flex items-center gap-2">
          <i class="pi pi-users text-blue-500"></i>
          <span>Cari Kartları İçe Aktar</span>
        </div>
      </template>
      <template #content>
        <p class="text-slate-600 mb-4">
          Excel (.xlsx) dosyasını kullanarak toplu cari kartı yükleyebilirsiniz. Mükerrer olan kayıtlar (Cari Kodu veya Vergi No) atlanacaktır.
        </p>
        
        <div class="flex flex-col gap-4">
          <div>
            <Button 
              label="Örnek Şablonu İndir" 
              icon="pi pi-download" 
              class="p-button-outlined p-button-secondary" 
              @click="handleDownloadSampleCari" 
              :loading="downloadingCari"
            />
          </div>
          <div class="border rounded p-4 bg-slate-50 border-slate-200">
            <FileUpload 
              mode="basic" 
              name="file" 
              accept=".xlsx" 
              :maxFileSize="10000000" 
              @uploader="customUploadCari" 
              customUpload 
              auto 
              chooseLabel="Excel Dosyası Seç (.xlsx)" 
              class="p-button-primary"
            />
          </div>
        </div>
      </template>
    </Card>

    <Card>
      <template #title>
        <div class="flex items-center gap-2">
          <i class="pi pi-box text-blue-500"></i>
          <span>Stok Kartları İçe Aktar</span>
        </div>
      </template>
      <template #content>
        <p class="text-slate-600 mb-4">
          Excel (.xlsx) dosyasını kullanarak toplu stok/ürün/hizmet kartı yükleyebilirsiniz. Mükerrer olan kayıtlar (Stok Kodu veya Barkod) atlanacaktır.
        </p>
        
        <div class="flex flex-col gap-4">
          <div>
            <Button 
              label="Örnek Şablonu İndir" 
              icon="pi pi-download" 
              class="p-button-outlined p-button-secondary" 
              @click="handleDownloadSampleProduct" 
              :loading="downloadingProduct"
            />
          </div>
          <div class="border rounded p-4 bg-slate-50 border-slate-200">
            <FileUpload 
              mode="basic" 
              name="file" 
              accept=".xlsx" 
              :maxFileSize="10000000" 
              @uploader="customUploadProduct" 
              customUpload 
              auto 
              chooseLabel="Excel Dosyası Seç (.xlsx)" 
              class="p-button-primary"
            />
          </div>
        </div>
      </template>
    </Card>
  </div>
</template>
