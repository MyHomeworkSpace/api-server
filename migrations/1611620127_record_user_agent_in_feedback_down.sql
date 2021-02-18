-- Description: Record user agent in feedback
-- Down migration
ALTER TABLE `feedback` DROP COLUMN `userAgent`;