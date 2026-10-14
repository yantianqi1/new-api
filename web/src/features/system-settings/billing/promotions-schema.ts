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
import { z } from 'zod'

const specialTierSchema = z.object({
  id: z.string().optional(),
  name: z.string().trim().min(1),
  price: z.number().min(0.01),
  quota: z.number().int().positive().max(Number.MAX_SAFE_INTEGER),
  total_limit: z.number().int().min(0).optional(),
  user_limit: z.number().int().min(0).optional(),
  description: z.string().optional(),
  enabled: z.boolean().optional(),
  sort_order: z.number().int().optional(),
  show_stock: z.boolean().optional(),
})

export const promotionsSchema = z.object({
  payment_setting_special_tiers: z.string().superRefine((value, context) => {
    try {
      if (
        z.array(specialTierSchema).safeParse(JSON.parse(value || '[]')).success
      ) {
        return
      }
    } catch {
      // The editor keeps invalid text so the administrator can correct it.
    }
    context.addIssue({
      code: 'custom',
      message: 'Invalid promotion tiers JSON',
    })
  }),
  bargain_setting_enabled: z.boolean(),
  bargain_setting_title: z.string(),
  bargain_setting_description: z.string(),
  bargain_setting_origin_price: z.number().min(0.01),
  bargain_setting_min_price: z.number().min(0.01),
  bargain_setting_cut_amount: z.number().positive(),
  bargain_setting_max_slashes: z.number().int().min(1),
  bargain_setting_quota: z
    .number()
    .int()
    .positive()
    .max(Number.MAX_SAFE_INTEGER),
  bargain_setting_user_limit: z.number().int().min(0),
  bargain_setting_ip_limit_hours: z.number().int().min(0),
})

export type PromotionsSettings = z.infer<typeof promotionsSchema>

export const PROMOTIONS_DEFAULT_VALUES: PromotionsSettings = {
  payment_setting_special_tiers: '[]',
  bargain_setting_enabled: false,
  bargain_setting_title: '',
  bargain_setting_description: '',
  bargain_setting_origin_price: 5,
  bargain_setting_min_price: 2,
  bargain_setting_cut_amount: 0.5,
  bargain_setting_max_slashes: 6,
  bargain_setting_quota: 500000,
  bargain_setting_user_limit: 1,
  bargain_setting_ip_limit_hours: 24,
}

export type PromotionsOptions = {
  [
    K in keyof PromotionsSettings as K extends `${infer Namespace}_setting_${infer Option}`
      ? `${Namespace}_setting.${Option}`
      : never
  ]: PromotionsSettings[K]
}

export const PROMOTIONS_DEFAULT_OPTIONS = Object.fromEntries(
  Object.entries(PROMOTIONS_DEFAULT_VALUES).map(([key, value]) => [
    key.replace('_setting_', '_setting.'),
    value,
  ])
) as PromotionsOptions
