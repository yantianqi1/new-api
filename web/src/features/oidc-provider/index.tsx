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
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { CopyButton } from '@/components/copy-button'
import { StaticDataTable } from '@/components/data-table'
import { Dialog } from '@/components/dialog'
import { EmptyState } from '@/components/empty-state'
import { ErrorState } from '@/components/error-state'
import { LoadingState } from '@/components/loading-state'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  SecureVerificationDialog,
  useSecureVerification,
  type RequestVerificationOptions,
} from '@/features/auth/secure-verification'
import { handleServerError } from '@/lib/handle-server-error'
import { AuthOperationError } from '@/lib/secure-verification'
import { useAuthStore } from '@/stores/auth-store'

import { SettingsSection } from '../system-settings/components/settings-section'
import {
  listOIDCClients,
  mutateOIDCClient,
  type OIDCClient,
  type OIDCClientInput,
} from './api'
import { OIDCClientFormDialog } from './client-form-dialog'

export function OIDCProviderSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const userId = useAuthStore((state) => state.auth.user?.id)
  const sessionId = useAuthStore((state) => state.auth.session?.sid)
  const verification = useSecureVerification()
  const operation = useRef<AbortController | null>(null)
  const [pending, setPending] = useState(false)
  const [editor, setEditor] = useState<{ client: OIDCClient | null } | null>(
    null
  )
  const [confirmation, setConfirmation] = useState<{
    client: OIDCClient
    action: 'delete' | 'secret.rotate'
  } | null>(null)
  const [secret, setSecret] = useState<{
    value: string
    userId: number | undefined
    sessionId: string | undefined
  } | null>(null)
  const list = useQuery({
    queryKey: ['oidc-provider-clients', userId],
    queryFn: listOIDCClients,
    retry: false,
  })
  const cancelVerification = verification.cancel
  useEffect(
    () => () => {
      operation.current?.abort()
      operation.current = null
      cancelVerification()
      setPending(false)
      setSecret(null)
      setEditor(null)
      setConfirmation(null)
    },
    [userId, sessionId, cancelVerification]
  )

  const run = async (
    action: 'create' | 'update' | 'delete' | 'secret.rotate',
    client?: OIDCClient,
    input?: OIDCClientInput
  ) => {
    if (operation.current || (action !== 'create' && !client)) return
    const controller = new AbortController()
    operation.current = controller
    setPending(true)
    const request: RequestVerificationOptions =
      action === 'create'
        ? { scope: 'oidc.client.create', context: {} }
        : {
            scope: `oidc.client.${action}`,
            context: { client_id: client?.id ?? 0 },
          }
    try {
      const proof = await verification.requestVerification(request)
      if (!proof || operation.current !== controller) return
      const result = await mutateOIDCClient(
        action,
        client?.id,
        input,
        proof.proof_token,
        controller.signal
      )
      if (operation.current !== controller) return
      setEditor(null)
      setConfirmation(null)
      if (result?.client_secret) {
        setSecret({ value: result.client_secret, userId, sessionId })
      }
      void queryClient.invalidateQueries({
        queryKey: ['oidc-provider-clients', userId],
      })
    } catch (error) {
      if (operation.current !== controller) return
      const failure = AuthOperationError.from(error)
      if (failure.code !== 'AUTH_CANCELLED') handleServerError(failure)
    } finally {
      if (operation.current === controller) {
        operation.current = null
        setPending(false)
      }
    }
  }
  const visibleSecret =
    secret && secret.userId === userId && secret.sessionId === sessionId
      ? secret.value
      : null
  return (
    <SettingsSection title={t('OIDC Provider')}>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Register applications that sign in with accounts from this gateway.'
        )}
      </p>
      <div>
        <Button disabled={pending} onClick={() => setEditor({ client: null })}>
          {t('Create OIDC application')}
        </Button>
      </div>
      {list.isPending && <LoadingState />}
      {list.isError && <ErrorState onRetry={() => void list.refetch()} />}
      {list.data && !list.data.length && (
        <EmptyState title={t('No OIDC applications')} />
      )}
      {!!list.data?.length && (
        <StaticDataTable>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Application name')}</TableHead>
              <TableHead>{t('Client ID')}</TableHead>
              <TableHead>{t('Redirect URIs')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead>{t('Actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {list.data.map((client) => (
              <TableRow key={client.id}>
                <TableCell>{client.name}</TableCell>
                <TableCell>
                  <div className='flex items-center gap-2'>
                    <code>{client.client_id}</code>
                    <CopyButton value={client.client_id} />
                  </div>
                </TableCell>
                <TableCell className='max-w-xs break-all whitespace-pre-wrap'>
                  {client.redirect_uris}
                </TableCell>
                <TableCell>
                  {t(client.enabled ? 'Enabled' : 'Disabled')}
                  {client.trusted && (
                    <span className='ml-2'>{t('Trusted')}</span>
                  )}
                </TableCell>
                <TableCell>
                  <div className='flex gap-2'>
                    <Button
                      size='sm'
                      variant='outline'
                      disabled={pending}
                      onClick={() => setEditor({ client })}
                    >
                      {t('Edit')}
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      disabled={pending}
                      onClick={() =>
                        setConfirmation({ client, action: 'secret.rotate' })
                      }
                    >
                      {t('Rotate client secret')}
                    </Button>
                    <Button
                      size='sm'
                      variant='destructive'
                      disabled={pending}
                      onClick={() =>
                        setConfirmation({ client, action: 'delete' })
                      }
                    >
                      {t('Delete')}
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </StaticDataTable>
      )}
      {editor && (
        <OIDCClientFormDialog
          key={editor.client?.id ?? 'new'}
          client={editor.client}
          pending={pending}
          hidden={verification.isActive}
          onClose={() => setEditor(null)}
          onSave={(input) =>
            run(
              editor.client ? 'update' : 'create',
              editor.client ?? undefined,
              input
            )
          }
        />
      )}
      <ConfirmDialog
        open={Boolean(confirmation) && !verification.isActive}
        onOpenChange={(open) => {
          if (!open && !pending) setConfirmation(null)
        }}
        title={t(
          confirmation?.action === 'delete'
            ? 'Delete OIDC application'
            : 'Rotate client secret'
        )}
        desc={t(
          confirmation?.action === 'delete'
            ? 'This application will no longer be able to sign in.'
            : 'The previous client secret will stop working immediately.'
        )}
        destructive
        disabled={pending}
        isLoading={pending}
        handleConfirm={() => {
          if (confirmation) void run(confirmation.action, confirmation.client)
        }}
      />
      {visibleSecret && (
        <Dialog
          open
          onOpenChange={(open) => {
            if (!open) setSecret(null)
          }}
          title={t('Client Secret')}
          description={t('Copy this secret now. It will not be shown again.')}
          footer={<Button onClick={() => setSecret(null)}>{t('Close')}</Button>}
        >
          <Label htmlFor='generated-oidc-secret'>{t('Client Secret')}</Label>
          <div className='flex gap-2'>
            <Input
              id='generated-oidc-secret'
              value={visibleSecret}
              readOnly
              autoComplete='off'
            />
            <CopyButton
              value={visibleSecret}
              aria-label={t('Copy client secret')}
            />
          </div>
        </Dialog>
      )}
      <SecureVerificationDialog {...verification.dialogProps} />
    </SettingsSection>
  )
}
