-- Create Instance Groups for each Arr type
INSERT INTO instance_groups (id, app_type, name, mode) VALUES
  (gen_random_uuid(), 'sonarr', 'Default', 'quality_hierarchy'),
  (gen_random_uuid(), 'radarr', 'Default', 'quality_hierarchy'),
  (gen_random_uuid(), 'lidarr', 'Default', 'quality_hierarchy'),
  (gen_random_uuid(), 'readarr', 'Default', 'quality_hierarchy'),
  (gen_random_uuid(), 'whisparr', 'Default', 'quality_hierarchy');

-- Add Sonarr instances to group
WITH sonarr_group AS (
  SELECT id FROM instance_groups WHERE app_type='sonarr' AND name='Default'
),
sonarr_instances AS (
  SELECT id FROM app_instances WHERE app_type='sonarr' ORDER BY name
)
INSERT INTO instance_group_members (group_id, instance_id, quality_rank, is_independent)
SELECT sg.id, si.id, ROW_NUMBER() OVER (ORDER BY si.id), false
FROM sonarr_group sg, sonarr_instances si;

-- Add Radarr instances to group
WITH radarr_group AS (
  SELECT id FROM instance_groups WHERE app_type='radarr' AND name='Default'
),
radarr_instances AS (
  SELECT id FROM app_instances WHERE app_type='radarr' ORDER BY name
)
INSERT INTO instance_group_members (group_id, instance_id, quality_rank, is_independent)
SELECT rg.id, ri.id, ROW_NUMBER() OVER (ORDER BY ri.id), false
FROM radarr_group rg, radarr_instances ri;

-- Add Lidarr instances
WITH lidarr_group AS (
  SELECT id FROM instance_groups WHERE app_type='lidarr' AND name='Default'
),
lidarr_instances AS (
  SELECT id FROM app_instances WHERE app_type='lidarr'
)
INSERT INTO instance_group_members (group_id, instance_id, quality_rank, is_independent)
SELECT lg.id, li.id, 1, true
FROM lidarr_group lg, lidarr_instances li;

-- Add Readarr instances
WITH readarr_group AS (
  SELECT id FROM instance_groups WHERE app_type='readarr' AND name='Default'
),
readarr_instances AS (
  SELECT id FROM app_instances WHERE app_type='readarr'
)
INSERT INTO instance_group_members (group_id, instance_id, quality_rank, is_independent)
SELECT rg.id, ri.id, 1, true
FROM readarr_group rg, readarr_instances ri;

-- Add Whisparr instances
WITH whisparr_group AS (
  SELECT id FROM instance_groups WHERE app_type='whisparr' AND name='Default'
),
whisparr_instances AS (
  SELECT id FROM app_instances WHERE app_type='whisparr'
)
INSERT INTO instance_group_members (group_id, instance_id, quality_rank, is_independent)
SELECT wg.id, wi.id, ROW_NUMBER() OVER (ORDER BY wi.id), false
FROM whisparr_group wg, whisparr_instances wi;

SELECT 'Instance groups and members created!' as status;
