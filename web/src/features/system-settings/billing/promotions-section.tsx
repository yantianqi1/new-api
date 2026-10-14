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

import { JsonCodeEditor } from '@/components/json-code-editor'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { formatJsonForEditor } from '../content/utils'
import { useUpdateOption } from '../hooks/use-update-option'
import {
  promotionsSchema,
  type PromotionsSettings,
  type PromotionsOptions,
} from './promotions-schema'

const numberFields = [
  {
    name: 'bargain_setting_origin_price',
    label: 'Original price',
    min: 0.01,
    step: 0.01,
  },
  {
    name: 'bargain_setting_min_price',
    label: 'Minimum price',
    min: 0.01,
    step: 0.01,
  },
  {
    name: 'bargain_setting_cut_amount',
    label: 'Discount per referral',
    min: 0.01,
    step: 0.01,
  },
  {
    name: 'bargain_setting_max_slashes',
    label: 'Maximum referrals',
    min: 1,
    step: 1,
  },
  { name: 'bargain_setting_quota', label: 'Package quota', min: 1, step: 1 },
  {
    name: 'bargain_setting_user_limit',
    label: 'Purchases per user',
    min: 0,
    step: 1,
  },
  {
    name: 'bargain_setting_ip_limit_hours',
    label: 'IP restriction window (hours)',
    min: 0,
    step: 1,
  },
] as const

export function PromotionsSection(props: { defaultValues: PromotionsOptions }) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const form = useForm<PromotionsSettings>({
    resolver: zodResolver(promotionsSchema),
    defaultValues: Object.fromEntries(
      Object.entries(props.defaultValues)
        .filter(
          ([key]) =>
            key.startsWith('bargain_setting.') ||
            key === 'payment_setting.special_tiers'
        )
        .map(([key, value]) => [key.replace('.', '_'), value])
    ) as PromotionsSettings,
  })
  const disabled = updateOption.isPending || form.formState.isSubmitting

  async function onSubmit(values: PromotionsSettings) {
    for (const key of Object.keys(values) as Array<keyof PromotionsSettings>) {
      if (values[key] === form.formState.defaultValues?.[key]) continue
      const value =
        key === 'payment_setting_special_tiers'
          ? JSON.stringify(JSON.parse(String(values[key]) || '[]'))
          : String(values[key])
      await updateOption.mutateAsync({
        key: key.replace('_setting_', '_setting.'),
        value,
      })
    }
    form.reset(values)
  }

  return (
    <SettingsSection title={t('Promotions')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={disabled}
            isSaveDisabled={!form.formState.isDirty}
          />
          <FormField
            control={form.control}
            name='payment_setting_special_tiers'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Promotion tiers JSON')}</FormLabel>
                <FormControl>
                  <JsonCodeEditor
                    value={formatJsonForEditor(field.value)}
                    onChange={field.onChange}
                    disabled={disabled}
                  />
                </FormControl>
                <FormDescription>
                  {t(
                    'Configure tier price, gifted quota, purchase limits and stock visibility. Quota uses internal quota units.'
                  )}
                </FormDescription>
                <FormMessage>
                  {form.formState.errors.payment_setting_special_tiers &&
                    t('Invalid promotion tiers JSON')}
                </FormMessage>
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='bargain_setting_enabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable bargain package')}</FormLabel>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                    disabled={disabled}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />
          {(
            ['bargain_setting_title', 'bargain_setting_description'] as const
          ).map((name) => (
            <FormField
              key={name}
              control={form.control}
              name={name}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>
                    {name === 'bargain_setting_title'
                      ? t('Package title')
                      : t('Package description')}
                  </FormLabel>
                  <FormControl>
                    <Input {...field} disabled={disabled} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
          {numberFields.map((option) => (
            <FormField
              key={option.name}
              control={form.control}
              name={option.name}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t(option.label)}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={option.min}
                      step={option.step}
                      value={field.value}
                      onBlur={field.onBlur}
                      ref={field.ref}
                      name={field.name}
                      onChange={(event) =>
                        field.onChange(Number(event.target.value))
                      }
                      disabled={disabled}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
