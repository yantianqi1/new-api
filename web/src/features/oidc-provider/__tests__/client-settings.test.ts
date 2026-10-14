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
import { describe, expect, test } from 'vitest'

import { oidcClientFormSchema, oidcClientPayload } from '../schema'

const form = {
  name: 'Example',
  redirectUris: 'https://app.example.com/callback',
  scopes: 'openid profile email',
  trusted: false,
  enabled: true,
}
describe('OIDC provider client configuration', () => {
  test('valid callbacks preserve exact destinations and least-privilege settings', () => {
    expect(oidcClientPayload(oidcClientFormSchema.parse(form))).toEqual({
      name: 'Example',
      redirect_uris: ['https://app.example.com/callback'],
      scopes: ['openid', 'profile', 'email'],
      trusted: false,
      enabled: true,
    })
  })
  test.each([
    'http://app.example.com/callback',
    'https://app.example.com/*',
    'https://app.example.com/callback#fragment',
    'javascript:alert(1)',
  ])('unsafe redirect %s is rejected', (uri) => {
    expect(
      oidcClientFormSchema.safeParse({ ...form, redirectUris: uri }).success
    ).toBe(false)
  })
  test('HTTP development callbacks are allowed only on loopback addresses', () => {
    expect(
      oidcClientFormSchema.safeParse({
        ...form,
        redirectUris: 'http://localhost:3000/callback',
      }).success
    ).toBe(true)
  })
  test.each(['profile email', 'openid admin'])(
    'invalid grant %s is rejected',
    (scopes) => {
      expect(oidcClientFormSchema.safeParse({ ...form, scopes }).success).toBe(
        false
      )
    }
  )
})
