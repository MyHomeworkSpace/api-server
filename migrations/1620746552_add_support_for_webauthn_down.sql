-- Description: add support for Webauthn
-- Down migration

ALTER TABLE `2fa` DROP COLUMN `webauthn`;
ALTER TABLE `2fa` CHANGE `totp` `secret` TEXT;
ALTER TABLE `2fa` RENAME TO `totp`;