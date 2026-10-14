/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { ErrorState } from '@/components/error-state'
import { LoadingState } from '@/components/loading-state'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { api } from '@/lib/api'
import { authRequestOptions, authResult } from '@/lib/secure-verification'
import { getServerErrorMessage } from '@/lib/server-error-message'
import { useAuthStore } from '@/stores/auth-store'

interface OIDCAuthorizationResult {
  authorization_url?: string
  consent_token?: string
  client_name?: string
  scope?: string
  expires_in?: number
}

function followOIDCAuthorization(result: OIDCAuthorizationResult) {
  if (!result.authorization_url) return false
  const url = new URL(result.authorization_url)
  if (
    url.protocol !== 'https:' &&
    !(
      url.protocol === 'http:' &&
      ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname)
    )
  ) {
    throw new Error('Invalid authorization redirect')
  }
  window.location.assign(url.href)
  return true
}

export function OIDCAuthorize() {
  const { t } = useTranslation()
  const userId = useAuthStore((state) => state.auth.user?.id)
  const sessionId = useAuthStore((state) => state.auth.session?.sid)
  const [received, setResult] = useState<{
    data: OIDCAuthorizationResult
    userId: number | undefined
    sessionId: string | undefined
  } | null>(null)
  const result =
    received && received.userId === userId && received.sessionId === sessionId
      ? received.data
      : null
  const decisionRequest = useRef<AbortController | null>(null)
  const [failure, setFailure] = useState<string | null>(null)
  const [pending, setPending] = useState(false)
  useEffect(() => {
    const controller = new AbortController()
    void authResult<OIDCAuthorizationResult>(
      api.get(`/oauth/authorize${window.location.search}`, {
        ...authRequestOptions,
        headers: { Accept: 'application/json' },
        singleUseAuthorization: true,
        disableDuplicate: true,
        signal: controller.signal,
      })
    )
      .then((response) => {
        if (controller.signal.aborted) return
        if (!followOIDCAuthorization(response)) {
          if (!response.consent_token) {
            throw new Error('Invalid authorization response')
          }
          setFailure(null)
          setResult({ data: response, userId, sessionId })
        }
      })
      .catch((error) => {
        if (!controller.signal.aborted) setFailure(getServerErrorMessage(error))
      })
    return () => {
      controller.abort()
      decisionRequest.current?.abort()
      decisionRequest.current = null
    }
  }, [userId, sessionId])

  const decide = async (decision: 'allow' | 'deny') => {
    if (!result?.consent_token || pending || decisionRequest.current) return
    const controller = new AbortController()
    decisionRequest.current = controller
    setPending(true)
    try {
      const response = await authResult<OIDCAuthorizationResult>(
        api.post(
          '/oauth/authorize',
          { consent_token: result.consent_token, decision },
          {
            ...authRequestOptions,
            singleUseAuthorization: true,
            signal: controller.signal,
          }
        )
      )
      if (
        controller.signal.aborted ||
        useAuthStore.getState().auth.session?.sid !== sessionId
      ) {
        return
      }
      if (!followOIDCAuthorization(response)) {
        throw new Error('Invalid authorization response')
      }
    } catch (error) {
      if (controller.signal.aborted) return
      setResult(null)
      setFailure(getServerErrorMessage(error))
    } finally {
      if (decisionRequest.current === controller) {
        decisionRequest.current = null
        setPending(false)
      }
    }
  }

  return (
    <main className='bg-background flex min-h-svh items-center justify-center p-4'>
      {failure && (
        <ErrorState
          title={t('Authorization failed')}
          description={t(failure)}
        />
      )}
      {!failure && !result && <LoadingState />}
      {!failure && result && (
        <Card className='w-full max-w-md'>
          <CardHeader>
            <CardTitle>{t('Authorize application')}</CardTitle>
            <CardDescription>
              {t(
                'Allow this application to sign in with your gateway account?'
              )}
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <p className='font-semibold break-words'>{result.client_name}</p>
            <p className='text-muted-foreground text-sm'>
              {t('Scopes')}: {result.scope}
            </p>
            {result.expires_in && (
              <p className='text-muted-foreground text-sm'>
                {t('Authorization lasts {{minutes}} minutes', {
                  minutes: Math.ceil(result.expires_in / 60),
                })}
              </p>
            )}
            <div className='flex gap-2'>
              <Button
                className='flex-1'
                variant='outline'
                disabled={pending}
                onClick={() => void decide('deny')}
              >
                {t('Deny')}
              </Button>
              <Button
                className='flex-1'
                disabled={pending}
                onClick={() => void decide('allow')}
              >
                {t('Authorize')}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </main>
  )
}
