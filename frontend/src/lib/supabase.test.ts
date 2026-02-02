import { describe, it, expect, vi, beforeEach } from 'vitest';

// Store original env
const originalEnv = process.env;

describe('supabase client', () => {
  beforeEach(() => {
    // Reset modules to clear cached supabase instance
    vi.resetModules();
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('returns null when NEXT_PUBLIC_SUPABASE_URL is not set', async () => {
    process.env.NEXT_PUBLIC_SUPABASE_URL = '';
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = 'test-key';

    const { getSupabase } = await import('./supabase');
    const client = getSupabase();

    expect(client).toBeNull();
  });

  it('returns null when NEXT_PUBLIC_SUPABASE_ANON_KEY is not set', async () => {
    process.env.NEXT_PUBLIC_SUPABASE_URL = 'https://test.supabase.co';
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = '';

    const { getSupabase } = await import('./supabase');
    const client = getSupabase();

    expect(client).toBeNull();
  });

  it('creates client when credentials are provided', async () => {
    process.env.NEXT_PUBLIC_SUPABASE_URL = 'https://test.supabase.co';
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = 'test-anon-key';

    const { getSupabase } = await import('./supabase');
    const client = getSupabase();

    expect(client).not.toBeNull();
    expect(client).toHaveProperty('channel');
    expect(client).toHaveProperty('removeChannel');
  });

  it('returns singleton instance', async () => {
    process.env.NEXT_PUBLIC_SUPABASE_URL = 'https://test.supabase.co';
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = 'test-anon-key';

    const { getSupabase } = await import('./supabase');
    const client1 = getSupabase();
    const client2 = getSupabase();

    expect(client1).toBe(client2);
  });
});
