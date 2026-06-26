import { useAuthStore } from '@/stores/auth'

type Action = 'create' | 'read' | 'update' | 'delete'

export function usePermission() {
  const authStore = useAuthStore()

  function can(module: string, action: Action): boolean {
    const role = authStore.user?.role
    if (role === 'admin' || role === 'superadmin') return true
    if (role !== 'personel') return false

    const perm = authStore.permissions.find((p) => p.module === module)
    if (!perm) return false

    switch (action) {
      case 'create': return perm.can_create
      case 'read': return perm.can_read
      case 'update': return perm.can_update
      case 'delete': return perm.can_delete
      default: return false
    }
  }

  return { can }
}
