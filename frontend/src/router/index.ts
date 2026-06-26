import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AuthLayout from '@/layouts/AuthLayout.vue'
import AppLayout from '@/layouts/AppLayout.vue'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

// NProgress Configuration
NProgress.configure({ showSpinner: false, speed: 400, minimum: 0.1 })

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      path: '/auth',
      component: AuthLayout,
      children: [
        {
          path: '/login',
          component: () => import('@/views/auth/Login.vue'),
        },
        {
          path: '/register',
          component: () => import('@/views/auth/Register.vue'),
        },
      ],
    },
    {
      path: '/pay/:token',
      name: 'PublicInvoice',
      component: () => import('@/views/invoice/PublicView.vue'),
      meta: { public: true },
    },
    {
      path: '/quote/:token',
      name: 'PublicQuote',
      component: () => import('@/views/quote/PublicView.vue'),
      meta: { public: true },
    },
    {
      path: '/dashboard',
      component: AppLayout,
      meta: { requiresTenant: true },
      children: [
        {
          path: '',
          component: () => import('@/views/Dashboard.vue'),
        },
      ],
    },
    // Placeholder routes for standard modules
    {
      path: '/caris',
      component: AppLayout,
      meta: { module: 'caris' },
      children: [
        {
          path: '',
          component: () => import('@/views/cari/List.vue'),
        },
        {
          path: ':id',
          component: () => import('@/views/cari/Detail.vue'),
        },
      ],
    },
    {
      path: '/invoices',
      component: AppLayout,
      meta: { module: 'invoices' },
      children: [
        {
          path: '',
          component: () => import('@/views/invoice/List.vue'),
        },
        {
          path: 'new',
          component: () => import('@/views/invoice/Form.vue'),
        },
        {
          path: ':id',
          component: () => import('@/views/invoice/Detail.vue'),
        },
        {
          path: ':id/edit',
          component: () => import('@/views/invoice/Form.vue'),
        },
      ],
    },
    {
      path: '/quotes',
      component: AppLayout,
      meta: { module: 'invoices' },
      children: [
        {
          path: '',
          component: () => import('@/views/quote/List.vue'),
        },
        {
          path: 'new',
          component: () => import('@/views/quote/Form.vue'),
        },
        {
          path: ':id',
          component: () => import('@/views/quote/Detail.vue'),
        },
        {
          path: ':id/edit',
          component: () => import('@/views/quote/Form.vue'),
        },
      ],
    },
    {
      path: '/projects',
      component: AppLayout,
      meta: { module: 'invoices' },
      children: [
        {
          path: '',
          component: () => import('@/views/project/List.vue'),
        },
        {
          path: 'new',
          component: () => import('@/views/project/Form.vue'),
        },
        {
          path: ':id',
          component: () => import('@/views/project/Detail.vue'),
        },
        {
          path: ':id/edit',
          component: () => import('@/views/project/Form.vue'),
        },
      ],
    },
    {
      path: '/payments',
      component: AppLayout,
      meta: { module: 'payments' },
      children: [
        {
          path: '',
          name: 'Payments',
          component: () => import('@/views/payment/List.vue'),
        },
      ],
    },
    {
      path: '/cash',
      component: AppLayout,
      meta: { module: 'payments' },
      children: [
        {
          path: '',
          name: 'Cash',
          component: () => import('@/views/cash/List.vue'),
        },
      ],
    },
    {
      path: '/expenses',
      component: AppLayout,
      meta: { module: 'expenses' },
      children: [
        {
          path: '',
          component: () => import('@/views/expense/List.vue'),
        },
      ],
    },
    {
      path: '/products',
      component: AppLayout,
      meta: { module: 'products' },
      children: [
        {
          path: '',
          component: () => import('@/views/product/List.vue'),
        },
      ],
    },
    {
      path: '/reports',
      component: AppLayout,
      meta: { module: 'reports' },
      children: [
        {
          path: '',
          component: () => import('@/views/report/List.vue'),
        },
      ],
    },
    {
      path: '/employees',
      component: AppLayout,
      meta: { adminOnly: true },
      children: [
        {
          path: '',
          component: () => import('@/views/employee/List.vue'),
        },
      ],
    },
    {
      path: '/settings',
      component: AppLayout,
      children: [
        {
          path: '',
          component: () => import('@/views/settings/Index.vue'),
        },
        {
          path: 'modules',
          name: 'modul-ayarlari',
          component: () => import('@/views/settings/ModulAyarlari.vue'),
        },
      ],
    },
    {
      path: '/announcements',
      component: AppLayout,
      children: [
        {
          path: '',
          component: () => import('@/views/announcement/Board.vue'),
        },
      ],
    },
    {
      path: '/billing',
      redirect: (to) => ({ path: '/settings', query: { ...to.query, tab: 'billing' } })
    },
    {
      path: '/unauthorized',
      component: AppLayout,
      children: [
        {
          path: '',
          component: () => import('@/views/Unauthorized.vue'),
        },
      ],
    },
    // Superadmin Routes
    {
      path: '/superadmin',
      component: AppLayout,
      meta: { role: 'superadmin' },
      children: [
        {
          path: 'dashboard',
          component: () => import('@/views/superadmin/Dashboard.vue'),
        },
        {
          path: 'announcements',
          component: () => import('@/views/superadmin/Announcements.vue'),
        },
        {
          path: 'companies',
          component: () => import('@/views/superadmin/Companies.vue'),
        },
        {
          path: 'plans',
          component: () => import('@/views/superadmin/Plans.vue'),
        },
        {
          path: 'email-settings',
          component: () => import('@/views/superadmin/EmailSettings.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to, _, next) => {
  NProgress.start()
  const authStore = useAuthStore()

  const isAuthRoute = to.path === '/login' || to.path === '/register'
  const isPublicRoute = to.meta.public === true

  // Public sayfalar (örn. /pay/:token) oturum sorgusu yapmadan erişilebilir olmalı;
  // aksi halde 401 -> refresh -> login redirect zinciri ziyaretçiyi login'e atar.
  if (isPublicRoute) {
    next()
    return
  }

  // Fetch session on first load if not ready
  if (!authStore.ready) {
    await authStore.fetchMe()
  }

  if (authStore.isAuthenticated) {
    const isSuperAdmin = authStore.user?.role === 'superadmin'

    // Block non-superadmin users from superadmin pages
    if (to.meta.role === 'superadmin' && !isSuperAdmin) {
      next('/dashboard')
      return
    }

    if (isSuperAdmin) {
      if (to.path === '/dashboard' || to.path === '/' || isAuthRoute || to.path === '/superadmin') {
        next('/superadmin/dashboard')
        return
      }
      if (to.meta.requiresTenant) {
        next('/superadmin/dashboard')
        return
      }
      next()
      return
    }

    if (isAuthRoute) {
      next('/dashboard')
      return
    }

    // Enabled modules check for tenant users
    const company = authStore.company
    if (company && company.enabled_modules) {
      try {
        const enabledModules = JSON.parse(company.enabled_modules) as string[]
        // Dashboard ve Ayarlar kasıtlı olarak burada yok: her zaman erişilebilir
        // kalmalı, aksi halde admin modülleri kapatıp kendini kilitleyebilir.
        const pathMap: Record<string, string> = {
          '/caris': 'caris',
          '/invoices': 'invoices',
          '/quotes': 'quotes',
          '/projects': 'invoices',
          '/payments': 'payments',
          '/cash': 'cash',
          '/expenses': 'expenses',
          '/products': 'products',
          '/reports': 'reports',
          '/employees': 'employees'
        }

        // Find active mapping based on start of path
        const matchedPrefix = Object.keys(pathMap).find(p => to.path.startsWith(p))
        if (matchedPrefix) {
          const modKey = pathMap[matchedPrefix]
          if (enabledModules.length > 0 && !enabledModules.includes(modKey)) {
            next('/unauthorized')
            return
          }
        }
      } catch (e) {}
    }

    // Sadece yöneticiye açık sayfalar (örn. Personel yönetimi). Backend de
    // bu route'ları admin/superadmin ile sınırlar; burada erken engelliyoruz.
    if (to.meta.adminOnly && authStore.user?.role !== 'admin') {
      next('/unauthorized')
      return
    }

    // Modül bazlı sayfa erişim kontrolü (sadece 'personel' rolü için)
    if (authStore.user?.role === 'personel' && to.meta.module) {
      const perm = authStore.permissions.find((p) => p.module === to.meta.module)
      if (!perm || !perm.can_read) {
        next('/unauthorized')
        return
      }
    }

    // Subscription status validation
    const isTrialValid = company?.subscription_status === 'trial' && new Date(company.trial_ends_at) > new Date()
    const isActive = company?.subscription_status === 'active'
    const isBillingPath = to.path === '/billing' || (to.path === '/settings' && to.query.tab === 'billing')

    if (!isSuperAdmin && !isTrialValid && !isActive && !isBillingPath) {
      next('/billing')
    } else {
      next()
    }
  } else {
    if (isAuthRoute) {
      next()
    } else {
      next('/login')
    }
  }
})

router.afterEach(() => {
  NProgress.done()
})

export default router
