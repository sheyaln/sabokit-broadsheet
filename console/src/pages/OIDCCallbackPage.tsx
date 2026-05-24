import { useEffect, useRef } from 'react'
import { useNavigate, useSearch } from '@tanstack/react-router'
import { Spin, App } from 'antd'
import { useAuth } from '../contexts/AuthContext'
import { MainLayout } from '../layouts/MainLayout'
import { useLingui } from '@lingui/react/macro'

interface OIDCCallbackSearch {
  token?: string
  expires_at?: string
}

export function OIDCCallbackPage() {
  const { t } = useLingui()
  const { signin } = useAuth()
  const navigate = useNavigate()
  const { message } = App.useApp()
  const search = useSearch({ from: '/console/auth/oidc/callback' }) as OIDCCallbackSearch
  const hasProcessed = useRef(false)

  useEffect(() => {
    if (hasProcessed.current) return
    hasProcessed.current = true

    const { token } = search

    if (!token) {
      message.error(t`SSO authentication failed: no token received`)
      navigate({ to: '/console/signin', replace: true })
      return
    }

    signin(token)
      .then(() => {
        message.success(t`Successfully signed in via SSO`)
        setTimeout(() => {
          navigate({ to: '/console', replace: true })
        }, 100)
      })
      .catch(() => {
        message.error(t`SSO authentication failed`)
        navigate({ to: '/console/signin', replace: true })
      })
  }, [search, signin, navigate, message, t])

  return (
    <MainLayout>
      <div className="flex items-center justify-center h-[calc(100vh-48px)]">
        <Spin size="large" tip={t`Completing SSO sign-in...`}>
          <div style={{ padding: 50 }} />
        </Spin>
      </div>
    </MainLayout>
  )
}
