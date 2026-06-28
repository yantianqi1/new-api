import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { buildTokenGroupOptions } from './token-group-options.js';

describe('buildTokenGroupOptions', () => {
  it('uses the group name as the label when the backend description is empty', () => {
    const options = buildTokenGroupOptions({
      'GPT0.5': { desc: '', ratio: 0.5 },
    });

    assert.deepEqual(options, [
      {
        label: 'GPT0.5',
        value: 'GPT0.5',
        ratio: 0.5,
      },
    ]);
  });
});
