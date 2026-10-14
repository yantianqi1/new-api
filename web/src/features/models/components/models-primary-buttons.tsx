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
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, MoreHorizontal, List, AlertCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useCanEditModelPricing } from '@/features/model-pricing/api'
import { handleServerError } from '@/lib/handle-server-error'
import { requireServerSuccess } from '@/lib/server-error-message'

import { syncMarketplaceModels } from '../api'
import { useModels } from './models-provider'

export function ModelsPrimaryButtons() {
  const { t } = useTranslation()
  const { setOpen, setCurrentRow } = useModels()
  const canPrice = useCanEditModelPricing()
  const queryClient = useQueryClient()
  const marketplaceSync = useMutation({
    mutationFn: async () => requireServerSuccess(await syncMarketplaceModels()),
    onSuccess: async (response) => {
      const result = response.data
      if (result) {
        toast.success(
          t(
            'Model Square synced: {{created}} added, {{existing}} existing, {{total}} total.',
            {
              created: result.created_models,
              existing: result.existing_models,
              total: result.total_models,
            }
          )
        )
      }
      await Promise.all(
        ['models', 'pricing'].map((key) =>
          queryClient.invalidateQueries({ queryKey: [key] })
        )
      )
    },
    onError: (error) => handleServerError(error),
  })

  const handleCreateModel = () => {
    setCurrentRow(null)
    setOpen('create-model')
  }

  const handleMissingModels = () => {
    setOpen('missing-models')
  }

  const handleSync = () => {
    setOpen('sync-wizard')
  }

  const handlePrefillGroups = () => {
    setOpen('prefill-groups')
  }

  return (
    <div className='flex flex-wrap items-center gap-2'>
      <Button onClick={handleSync} variant='outline' size='sm'>
        {t('Sync metadata')}
      </Button>
      {canPrice && (
        <Button
          onClick={() => setOpen('price-sync')}
          variant='outline'
          size='sm'
        >
          {t('Sync pricing')}
        </Button>
      )}
      <Button
        variant='outline'
        size='sm'
        disabled={marketplaceSync.isPending}
        onClick={() => marketplaceSync.mutate()}
      >
        {t('Sync Model Square')}
      </Button>
      {/* Create Model */}
      <Button onClick={handleCreateModel} size='sm'>
        <Plus className='h-4 w-4' />
        {t('Add Model')}
      </Button>

      {/* More Actions */}
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant='outline' size='sm' aria-label={t('Open menu')} />
          }
        >
          <MoreHorizontal className='h-4 w-4' />
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' className='w-56'>
          <DropdownMenuItem onClick={handleMissingModels}>
            {t('Missing Models')}
            <DropdownMenuShortcut>
              <AlertCircle className='h-4 w-4' />
            </DropdownMenuShortcut>
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={handlePrefillGroups}>
            {t('Prefill Groups')}
            <DropdownMenuShortcut>
              <List className='h-4 w-4' />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
