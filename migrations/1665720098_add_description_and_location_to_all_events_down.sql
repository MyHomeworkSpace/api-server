-- Description: Add description and location to all events
-- Down migration

ALTER TABLE `calendar_events` DROP COLUMN `location`;

ALTER TABLE `calendar_hwevents` DROP COLUMN `desc`;
ALTER TABLE `calendar_hwevents` DROP COLUMN `location`;