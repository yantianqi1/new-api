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
import assert from 'node:assert/strict'
import test from 'node:test'

import type { UsageLog } from '../data/schema.ts'
import { formatModelName } from './model-format.ts'

function makeUsageLog(other: Record<string, unknown>): UsageLog {
  return {
    id: 1,
    user_id: 7,
    created_at: 1,
    type: 2,
    content: '',
    username: 'alice',
    token_name: 'default',
    model_name: '[free]gemini-3.1-pro-preview',
    quota: 0,
    prompt_tokens: 0,
    completion_tokens: 0,
    use_time: 0,
    is_stream: false,
    channel: 0,
    channel_name: '',
    token_id: 0,
    group: 'default',
    ip: '',
    other: JSON.stringify(other),
    request_id: '',
    upstream_request_id: '',
  }
}

test('hides mapped upstream model names from normal users', () => {
  const log = makeUsageLog({
    is_model_mapped: true,
    upstream_model_name: 'gemini-3.1-pro-preview',
  })

  assert.deepEqual(formatModelName(log, false), {
    name: '[free]gemini-3.1-pro-preview',
    isMapped: false,
    actualModel: undefined,
  })
})

test('keeps mapped upstream model names for admins', () => {
  const log = makeUsageLog({
    is_model_mapped: true,
    upstream_model_name: 'gemini-3.1-pro-preview',
  })

  assert.deepEqual(formatModelName(log, true), {
    name: '[free]gemini-3.1-pro-preview',
    isMapped: true,
    actualModel: 'gemini-3.1-pro-preview',
  })
})
