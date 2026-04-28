BEGIN;

INSERT INTO posts (
  id,
  author_id,
  content,
  created_at
) VALUES
  (
    'post-1',
    'user-1',
    'Shipping the first cut of the feed service today.',
    '2026-04-27T09:00:00Z'
  ),
  (
    'post-2',
    'user-2',
    'GraphQL is a lot nicer when the joins are explicit and cheap.',
    '2026-04-27T09:05:00Z'
  ),
  (
    'post-3',
    'user-3',
    'Observability before optimization still wins most weeks.',
    '2026-04-27T09:10:00Z'
  )
ON CONFLICT (id) DO UPDATE SET
  author_id = EXCLUDED.author_id,
  content = EXCLUDED.content,
  created_at = EXCLUDED.created_at;

INSERT INTO comments (
  id,
  post_id,
  author_id,
  content,
  created_at
) VALUES
  (
    'comment-1',
    'post-1',
    'user-2',
    'That should make the gateway much easier to test.',
    '2026-04-27T09:12:00Z'
  ),
  (
    'comment-2',
    'post-2',
    'user-1',
    'Fully agree. Batching is doing real work there.',
    '2026-04-27T09:15:00Z'
  )
ON CONFLICT (id) DO UPDATE SET
  post_id = EXCLUDED.post_id,
  author_id = EXCLUDED.author_id,
  content = EXCLUDED.content,
  created_at = EXCLUDED.created_at;

COMMIT;
