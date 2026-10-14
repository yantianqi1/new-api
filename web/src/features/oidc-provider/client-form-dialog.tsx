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
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

import type { OIDCClient, OIDCClientInput } from './api'
import {
  oidcClientFormSchema,
  oidcClientPayload,
  type OIDCClientFormValues,
} from './schema'

export function OIDCClientFormDialog(props: {
  client: OIDCClient | null
  hidden: boolean
  pending: boolean
  onClose: () => void
  onSave: (input: OIDCClientInput) => Promise<void>
}) {
  const { t } = useTranslation()
  const form = useForm<OIDCClientFormValues>({
    resolver: zodResolver(oidcClientFormSchema),
    defaultValues: {
      name: props.client?.name ?? '',
      redirectUris: props.client?.redirect_uris ?? '',
      scopes: props.client?.scopes ?? 'openid profile email',
      trusted: props.client?.trusted ?? false,
      enabled: props.client?.enabled ?? true,
    },
  })
  return (
    <Dialog
      open={!props.hidden}
      onOpenChange={(open) => {
        if (!open && !props.pending) props.onClose()
      }}
      title={t(
        props.client ? 'Edit OIDC application' : 'Create OIDC application'
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
            form='oidc-client-form'
            type='submit'
            disabled={props.pending}
          >
            {t('Save')}
          </Button>
        </>
      }
    >
      <Form {...form}>
        <form
          id='oidc-client-form'
          className='space-y-4'
          onSubmit={form.handleSubmit((values) =>
            props.onSave(oidcClientPayload(values))
          )}
        >
          <FormField
            control={form.control}
            name='name'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Application name')}</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='redirectUris'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Redirect URIs')}</FormLabel>
                <FormControl>
                  <Textarea {...field} rows={3} autoComplete='off' />
                </FormControl>
                <FormDescription>
                  {t(
                    'One exact redirect URI per line. HTTPS is required outside localhost.'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='scopes'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Scopes')}</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormDescription>
                  {t('Supported scopes: openid profile email')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='enabled'
            render={({ field }) => (
              <FormItem className='flex items-center justify-between'>
                <FormLabel>{t('Enabled')}</FormLabel>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='trusted'
            render={({ field }) => (
              <FormItem>
                <div className='flex items-center justify-between'>
                  <FormLabel>{t('Trusted application')}</FormLabel>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </div>
                <FormDescription>
                  {t(
                    'Trusted applications skip the user consent screen. Enable only for applications you control.'
                  )}
                </FormDescription>
              </FormItem>
            )}
          />
        </form>
      </Form>
    </Dialog>
  )
}
