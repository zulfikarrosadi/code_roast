CREATE TABLE IF NOT EXISTS `posts` (
  `id` varchar(36) NOT NULL,
  `caption` text,
  `user_id` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `status` enum('take_down','pending','published') DEFAULT 'published',
  `subforum_id` varchar(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `subforum_id` (`subforum_id`),
  CONSTRAINT `posts_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `posts_ibfk_2` FOREIGN KEY (`subforum_id`) REFERENCES `subforums` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci