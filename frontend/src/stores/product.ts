import { defineStore } from 'pinia'
import {
  getProductsApi,
  getProductByIDApi,
  createProductApi,
  updateProductApi,
  deleteProductApi,
  getProductCategoriesApi,
  createProductCategoryApi,
  updateProductCategoryApi,
  deleteProductCategoryApi,
  getWarehousesApi,
  createWarehouseApi,
  updateWarehouseApi,
  deleteWarehouseApi,
  createStockMovementApi,
  getProductMovementsApi,
} from '@/api/product'

export interface ProductCategory {
  id: string
  company_id: string
  name: string
  default_kdv_rate: string
}

export interface Warehouse {
  id: string
  company_id: string
  name: string
  address: string
  is_default: boolean
}

export interface Product {
  id: string
  company_id: string
  code: string
  name: string
  type: 'product' | 'service'
  unit: string
  barcode: string
  custom_codes: string
  brand?: string
  serial_numbers: string
  purchase_price: string
  average_cost: string
  sale_price: string
  currency: string
  tax_included: boolean
  purchase_tax_included: boolean
  tax_rate: string
  purchase_tax_rate: string
  description: string
  track_stock: boolean
  current_stock: string
  min_stock: string
  category_id: string | null
  category?: ProductCategory
  is_active: boolean
  created_at: string
}

export interface StockMovement {
  id: string
  company_id: string
  product_id: string
  product?: Product
  warehouse_id: string
  warehouse?: Warehouse
  date: string
  type: 'in' | 'out'
  source_type: 'invoice' | 'manual' | 'transfer'
  source_id: string | null
  quantity: string
  unit_cost: string
  balance_after: string
  note: string
  created_at: string
}

export const useProductStore = defineStore('product', {
  state: () => ({
    products: [] as Product[],
    total: 0,
    categories: [] as ProductCategory[],
    warehouses: [] as Warehouse[],
    movements: [] as StockMovement[],
    loading: false,
  }),
  actions: {
    async fetchProducts(params?: any) {
      this.loading = true
      try {
        const res = await getProductsApi(params)
        this.products = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchCategories() {
      const res = await getProductCategoriesApi()
      this.categories = res.data.data
    },
    async fetchWarehouses() {
      const res = await getWarehousesApi()
      this.warehouses = res.data.data
    },
    async fetchProductMovements(productId: string) {
      this.loading = true
      try {
        const res = await getProductMovementsApi(productId)
        this.movements = res.data.data
      } finally {
        this.loading = false
      }
    },
    async getProductByID(id: string) {
      const res = await getProductByIDApi(id)
      return res.data.data
    },
    async createProduct(data: any) {
      const res = await createProductApi(data)
      return res.data.data
    },
    async updateProduct(id: string, data: any) {
      const res = await updateProductApi(id, data)
      return res.data.data
    },
    async deleteProduct(id: string) {
      await deleteProductApi(id)
    },
    async createCategory(data: any) {
      const res = await createProductCategoryApi(data)
      await this.fetchCategories()
      return res.data.data
    },
    async updateCategory(id: string, data: any) {
      const res = await updateProductCategoryApi(id, data)
      await this.fetchCategories()
      return res.data.data
    },
    async deleteCategory(id: string) {
      await deleteProductCategoryApi(id)
      await this.fetchCategories()
    },
    async createWarehouse(data: any) {
      const res = await createWarehouseApi(data)
      await this.fetchWarehouses()
      return res.data.data
    },
    async updateWarehouse(id: string, data: any) {
      const res = await updateWarehouseApi(id, data)
      await this.fetchWarehouses()
      return res.data.data
    },
    async deleteWarehouse(id: string) {
      await deleteWarehouseApi(id)
      await this.fetchWarehouses()
    },
    async createStockMovement(data: any) {
      const res = await createStockMovementApi(data)
      return res.data.data
    },
  },
})
