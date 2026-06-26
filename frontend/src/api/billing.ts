import client from './client'

export const getBillingStatusApi = () => client.get('/billing/status')
export const getPlansApi = () => client.get('/billing/plans')
export const subscribeApi = (planID: string, periodType: string) => client.post('/billing/subscribe', { plan_id: planID, period_type: periodType })
export const simulateWebhookApi = (data: any) => client.post('/billing/webhook', data)
export const getTransactionsApi = () => client.get('/billing/transactions')
export const renewApi = (periodType: string) => client.post('/billing/renew', { period_type: periodType })

