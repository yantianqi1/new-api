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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ShieldCheck } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import {
  StaticDataTable,
  type StaticDataTableColumn,
} from '@/components/data-table'
import { EmptyState } from '@/components/empty-state'
import { ErrorState } from '@/components/error-state'
import { LoadingState } from '@/components/loading-state'
import { Button } from '@/components/ui/button'
import { TitledCard } from '@/components/ui/titled-card'
import { formatTimestampToDate } from '@/lib/format'
import { handleServerError } from '@/lib/handle-server-error'
import { useAuthStore } from '@/stores/auth-store'

import {
  listOIDCGrants,
  reduceOIDCGrantScope,
  revokeOIDCGrant,
  type OIDCGrant,
} from './api'
import { OIDCGrantPermissionsDialog } from './grant-permissions-dialog'

export function OIDCAuthorizedApplicationsCard() {
  const { t } = useTranslation()
  const userId = useAuthStore((state) => state.auth.user?.id)
  const queryClient = useQueryClient()
  const [selected, setSelected] = useState<{
    grant: OIDCGrant
    action: 'permissions' | 'revoke'
    userId: number | undefined
  } | null>(null)
  const target = selected?.userId === userId ? selected : null
  const queryKey = ['oidc-authorized-applications', userId]
  const grants = useQuery({
    queryKey,
    queryFn: ({ signal }) => listOIDCGrants(signal),
    enabled: userId !== undefined,
    retry: false,
  })
  const reduce = useMutation({
    mutationFn: ({ id, scope }: { id: number; scope: string }) =>
      reduceOIDCGrantScope(id, scope),
    retry: false,
    onSuccess: async () => {
      setSelected(null)
      toast.success(t('Application permissions updated'))
      await queryClient.invalidateQueries({ queryKey })
    },
    onError: (error) => handleServerError(error),
  })
  const revoke = useMutation({
    mutationFn: revokeOIDCGrant,
    retry: false,
    onSuccess: async () => {
      setSelected(null)
      toast.success(t('Application access revoked'))
      await queryClient.invalidateQueries({ queryKey })
    },
    onError: (error) => handleServerError(error),
  })
  const pending = reduce.isPending || revoke.isPending
  const columns = useMemo<StaticDataTableColumn<OIDCGrant>[]>(
    () => [
      {
        id: 'client_name',
        header: t('Application'),
        cell: (row) => (
          <div className='min-w-0'>
            <p className='font-medium break-words'>{row.client_name}</p>
            <code
              translate='no'
              className='text-muted-foreground text-xs break-all'
            >
              {row.client_id}
            </code>
          </div>
        ),
      },
      {
        id: 'scope',
        header: t('Permissions'),
        cell: (row) => (
          <code translate='no' className='text-xs break-words'>
            {row.scope}
          </code>
        ),
      },
      {
        id: 'expires_at',
        header: t('Expires at'),
        cell: (row) => (
          <span className='whitespace-nowrap'>
            {formatTimestampToDate(row.expires_at)}
          </span>
        ),
      },
      {
        id: 'actions',
        header: t('Actions'),
        cell: (row) => (
          <div className='flex flex-wrap gap-2'>
            <Button
              variant='outline'
              size='sm'
              disabled={pending}
              onClick={() =>
                setSelected({
                  grant: row,
                  action: 'permissions',
                  userId,
                })
              }
            >
              {t('Modify permissions')}
            </Button>
            <Button
              variant='outline'
              size='sm'
              disabled={pending}
              onClick={() =>
                setSelected({ grant: row, action: 'revoke', userId })
              }
            >
              {t('Revoke')}
            </Button>
          </div>
        ),
      },
    ],
    [t, pending, userId]
  )

  let content
  if (grants.isPending) content = <LoadingState />
  else if (grants.isError) {
    content = (
      <ErrorState
        description={t('Unable to load authorized applications')}
        onRetry={() => void grants.refetch()}
      />
    )
  } else if (!grants.data?.length) {
    content = (
      <EmptyState
        title={t('No authorized applications')}
        description={t(
          'Applications you authorize to use your account will appear here.'
        )}
        className='min-h-32'
      />
    )
  } else {
    content = (
      <StaticDataTable
        columns={columns}
        data={grants.data}
        getRowKey={(row) => String(row.id)}
      />
    )
  }

  return (
    <>
      <TitledCard
        title={t('OIDC Authorized Applications')}
        description={t(
          'Review application access, reduce permissions, or revoke authorization.'
        )}
        icon={<ShieldCheck className='size-4' />}
        disableHoverEffect
      >
        {content}
      </TitledCard>
      {target?.action === 'permissions' && (
        <OIDCGrantPermissionsDialog
          key={target.grant.id}
          grant={target.grant}
          pending={pending}
          onClose={() => setSelected(null)}
          onSave={async (scope) => {
            await reduce
              .mutateAsync({ id: target.grant.id, scope })
              .catch(() => undefined)
          }}
        />
      )}
      <ConfirmDialog
        open={target?.action === 'revoke'}
        onOpenChange={(open) => {
          if (!open && !pending) setSelected(null)
        }}
        title={t('Revoke application access?')}
        desc={t(
          'This application will no longer be able to access your account with this authorization.'
        )}
        confirmText={t('Revoke')}
        destructive
        isLoading={revoke.isPending}
        handleConfirm={() => {
          if (target?.action === 'revoke') revoke.mutate(target.grant.id)
        }}
      />
    </>
  )
}
