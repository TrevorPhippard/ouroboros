BEGIN;

INSERT INTO profiles (
  id,
  user_id,
  display_name,
  avatar_url,
  bio,
  headline,
  about,
  created_at
) VALUES
  (
    'profile-1',
    'user-1',
    'Alice Johnson',
    'https://api.dicebear.com/7.x/avataaars/svg?seed=user-1',
    'Backend engineer who likes tidy APIs and coffee.',
    'Senior Backend Engineer',
    'Building distributed systems and GraphQL APIs.',
    '2026-04-25T14:00:00Z'
  ),
  (
    'profile-2',
    'user-2',
    'Bob Smith',
    'https://api.dicebear.com/7.x/avataaars/svg?seed=user-2',
    'Product-minded engineer and frequent traveler.',
    'Staff Product Engineer',
    'Interested in product strategy, reliability, and developer experience.',
    '2026-04-25T14:05:00Z'
  ),
  (
    'profile-3',
    'user-3',
    'Carol Lee',
    'https://api.dicebear.com/7.x/avataaars/svg?seed=user-3',
    'Enjoys scaling services, observability, and good docs.',
    'Platform Engineer',
    'Focused on resilient infrastructure and internal platforms.',
    '2026-04-25T14:10:00Z'
  )
ON CONFLICT (id) DO UPDATE SET
  user_id = EXCLUDED.user_id,
  display_name = EXCLUDED.display_name,
  avatar_url = EXCLUDED.avatar_url,
  bio = EXCLUDED.bio,
  headline = EXCLUDED.headline,
  about = EXCLUDED.about;

COMMIT;
