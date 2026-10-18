export const clearExpiredUserSession = () => {
  if (typeof window === 'undefined') return

  localStorage.removeItem('user_token')
  localStorage.removeItem('user_profile')
  window.dispatchEvent(new Event('dujiao:user-auth-expired'))

  if (window.location.pathname.startsWith('/reseller')) {
    window.location.replace('/agent')
  }
}
