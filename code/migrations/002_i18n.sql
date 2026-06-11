-- Migration 002: i18n bilingual support
-- AC-I18N-003

ALTER TABLE blogs
  ADD COLUMN title_en           TEXT,
  ADD COLUMN body_en            TEXT,
  ADD COLUMN translation_status VARCHAR(20) NOT NULL DEFAULT 'none';

CREATE INDEX idx_blogs_translation_status ON blogs(translation_status)
  WHERE translation_status IN ('pending', 'failed');

ALTER TABLE users
  ADD COLUMN language_preference VARCHAR(5) NOT NULL DEFAULT 'vi';
