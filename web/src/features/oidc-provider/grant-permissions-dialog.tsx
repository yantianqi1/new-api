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
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form'

import type { OIDCGrant } from './api'

export function OIDCGrantPermissionsDialog(props: {
  grant: OIDCGrant
  pending: boolean
  onClose: () => void
  onSave: (scope: string) => Promise<void>
}) {
  const { t } = useTranslation()
  const grantedScopes = props.grant.scope.split(/\s+/).filter(Boolean)
  const form = useForm<{ scopes: string[] }>({
    defaultValues: { scopes: grantedScopes },
  })
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !props.pending) props.onClose()
      }}
      title={t('Modify permissions')}
      description={t(
        'Remove permissions previously granted to {{application}}. New permissions require a new authorization.',
        { application: props.grant.client_name }
      )}
      footer={
        <>
          <Button
            variant='outline'
            disabled={props.pending}
            onClick={props.onClose}
          >
            {t('Cancel')}
          </Button>
          <Button
            type='submit'
            form='oidc-grant-permissions'
            disabled={props.pending}
          >
            {t('Save changes')}
          </Button>
        </>
      }
    >
      <Form {...form}>
        <form
          id='oidc-grant-permissions'
          className='space-y-3'
          onSubmit={form.handleSubmit(async (values) => {
            const scope = grantedScopes
              .filter(
                (item) =>
                  !['profile', 'email'].includes(item) ||
                  values.scopes.includes(item)
              )
              .join(' ')
            await props.onSave(scope)
          })}
        >
          <FormField
            control={form.control}
            name='scopes'
            render={({ field }) => (
              <div className='space-y-3'>
                {grantedScopes.map((scope) => (
                  <FormItem key={scope} className='flex items-center gap-3'>
                    <FormControl>
                      <Checkbox
                        checked={
                          scope === 'openid' || field.value.includes(scope)
                        }
                        disabled={
                          props.pending || !['profile', 'email'].includes(scope)
                        }
                        onCheckedChange={(checked) => {
                          const remaining = field.value.filter(
                            (item) => item !== scope
                          )
                          field.onChange(
                            checked ? [...remaining, scope] : remaining
                          )
                        }}
                      />
                    </FormControl>
                    <FormLabel>
                      <code translate='no'>{scope}</code>
                    </FormLabel>
                  </FormItem>
                ))}
              </div>
            )}
          />
          <p className='text-muted-foreground text-sm'>
            {t(
              'openid is required for sign-in. Revoke the application to remove all access.'
            )}
          </p>
        </form>
      </Form>
    </Dialog>
  )
}
