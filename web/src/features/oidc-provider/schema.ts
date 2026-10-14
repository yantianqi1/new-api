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

export const oidcClientFormSchema = z.object({
  name: z.string().trim().min(1, 'Required'),
  redirectUris: z
    .string()
    .trim()
    .min(1, 'Required')
    .refine(
      (value) =>
        value
          .split(/\r?\n/)
          .filter(Boolean)
          .every((uri) => {
            try {
              const url = new URL(uri.trim())
              return (
                !url.hash &&
                !url.username &&
                !url.password &&
                !uri.includes('*') &&
                (url.protocol === 'https:' ||
                  (url.protocol === 'http:' &&
                    ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname)))
              )
            } catch {
              return false
            }
          }),
      'Redirect URIs must use HTTPS, or HTTP on localhost, without fragments or wildcards'
    ),
  scopes: z
    .string()
    .trim()
    .refine(
      (value) =>
        value.split(/\s+/).includes('openid') &&
        value
          .split(/\s+/)
          .every((scope) => ['openid', 'profile', 'email'].includes(scope)),
      'Select openid and only supported scopes'
    ),
  trusted: z.boolean(),
  enabled: z.boolean(),
})
export type OIDCClientFormValues = z.infer<typeof oidcClientFormSchema>

export function oidcClientPayload(values: OIDCClientFormValues) {
  return {
    name: values.name.trim(),
    redirect_uris: values.redirectUris
      .split(/\r?\n/)
      .map((uri) => uri.trim())
      .filter(Boolean),
    scopes: values.scopes.trim().split(/\s+/),
    trusted: values.trusted,
    enabled: values.enabled,
  }
}
