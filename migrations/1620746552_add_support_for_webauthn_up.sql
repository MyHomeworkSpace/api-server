-- Description: add support for Webauthn
-- Up migration
ALTER TABLE `totp` RENAME TO `2fa`;
ALTER TABLE `2fa` ADD `webauthn` varchar(255) NOT NULL DEFAULT '';
ALTER TABLE `2fa` CHANGE `secret` `totp` TEXT;
