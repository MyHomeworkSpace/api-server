-- Description: record user agent in feedback
-- Down migration
ALTER TABLE `myhomeworkspace`.`feedback` DROP COLUMN `userAgent`;