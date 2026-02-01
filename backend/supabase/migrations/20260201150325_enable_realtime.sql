-- SherlockOS Database Schema
-- Migration: 002_enable_realtime
-- Description: Enable Supabase Realtime for commits and jobs tables

-- Enable realtime for commits table (timeline updates)
ALTER PUBLICATION supabase_realtime ADD TABLE commits;

-- Enable realtime for jobs table (progress updates)
ALTER PUBLICATION supabase_realtime ADD TABLE jobs;

-- Note: scene_snapshots and suspect_profiles could also be added
-- if real-time updates for those are needed
-- ALTER PUBLICATION supabase_realtime ADD TABLE scene_snapshots;
-- ALTER PUBLICATION supabase_realtime ADD TABLE suspect_profiles;
