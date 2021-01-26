-- Description: record user agent in feedback
-- Up migration
ALTER TABLE feedback
ADD COLUMN userAgent varchar(255);
UPDATE feedback
SET userAgent = '';