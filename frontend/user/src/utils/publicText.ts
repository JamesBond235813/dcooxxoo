export const sanitizePublicText = (value: unknown, locale = 'zh-CN'): string => {
  const replacement = locale === 'en-US' ? 'overseas SMS service' : '海外短信服务'
  return String(value ?? '').replace(/5\s*sim(?:\.net)?/gi, replacement)
}
