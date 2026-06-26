import client from './client'

export const getProductsApi = (params?: any) => client.get('/products', { params })
export const getProductByIDApi = (id: string) => client.get(`/products/${id}`)
export const createProductApi = (data: any) => client.post('/products', data)
export const updateProductApi = (id: string, data: any) => client.put(`/products/${id}`, data)
export const deleteProductApi = (id: string) => client.delete(`/products/${id}`)
export const getCriticalStockApi = () => client.get('/products/critical-stock')
export const getNextProductCodeApi = () => client.get('/products/next-code')

// Categories
export const getProductCategoriesApi = () => client.get('/product-categories')
export const createProductCategoryApi = (data: any) => client.post('/product-categories', data)
export const updateProductCategoryApi = (id: string, data: any) => client.put(`/product-categories/${id}`, data)
export const deleteProductCategoryApi = (id: string) => client.delete(`/product-categories/${id}`)

// Warehouses
export const getWarehousesApi = () => client.get('/warehouses')
export const createWarehouseApi = (data: any) => client.post('/warehouses', data)
export const updateWarehouseApi = (id: string, data: any) => client.put(`/warehouses/${id}`, data)
export const deleteWarehouseApi = (id: string) => client.delete(`/warehouses/${id}`)

// Stock movements
export const createStockMovementApi = (data: any) => client.post('/stock-movements', data)
export const getProductMovementsApi = (productId: string) => client.get(`/products/${productId}/movements`)
