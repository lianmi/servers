-- MySQL dump 10.13  Distrib 8.0.18, for osx10.14 (x86_64)
--
-- Host: localhost    Database: lianmicloud
-- ------------------------------------------------------
-- Server version	8.0.19

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `contacts`
--

DROP TABLE IF EXISTS `contacts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contacts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `account` varchar(255) DEFAULT NULL,
  `alias` varchar(255) DEFAULT NULL,
  `source` varchar(255) DEFAULT NULL,
  `extend` varchar(255) DEFAULT NULL,
  `latest_chat_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `contacts`
--

LOCK TABLES `contacts` WRITE;
/*!40000 ALTER TABLE `contacts` DISABLE KEYS */;
/*!40000 ALTER TABLE `contacts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `devices`
--

DROP TABLE IF EXISTS `devices`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `devices` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `is_master` tinyint(1) DEFAULT NULL,
  `device_name` varchar(255) DEFAULT NULL,
  `device_index` int unsigned DEFAULT NULL,
  `os` varchar(255) DEFAULT NULL,
  `client_type` int unsigned DEFAULT NULL,
  `latest_logon_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `devices`
--

LOCK TABLES `devices` WRITE;
/*!40000 ALTER TABLE `devices` DISABLE KEYS */;
/*!40000 ALTER TABLE `devices` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `roles`
--

DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `roles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `user_name` varchar(255) DEFAULT NULL,
  `value` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `roles`
--

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` VALUES (1,1,'lsj001',''),(2,2,'admin','admin');
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tokens`
--

DROP TABLE IF EXISTS `tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tokens` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(255) DEFAULT NULL,
  `expired_at` timestamp NULL DEFAULT NULL,
  `token` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tokens`
--

LOCK TABLES `tokens` WRITE;
/*!40000 ALTER TABLE `tokens` DISABLE KEYS */;
INSERT INTO `tokens` VALUES (1,'lsj001','2020-08-03 10:13:23','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTY0NDk2MDIsIm9yaWdfaWF0IjoxNTk2NDQ2MDAyLCJ1c2VyTmFtZSI6ImxzajAwMSIsInVzZXJSb2xlcyI6Ilt7XCJpZFwiOjEsXCJ1c2VyX2lkXCI6MSxcInVzZXJfbmFtZVwiOlwibHNqMDAxXCIsXCJ2YWx1ZVwiOlwiXCJ9XSJ9.8awY448AkEkUSZnLL9W4eXpoRJzQlCPVM3ht1Z0gjjw'),(2,'lsj001','2020-08-03 10:24:22','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTY0NTAyNjEsIm9yaWdfaWF0IjoxNTk2NDQ2NjYxLCJ1c2VyTmFtZSI6ImxzajAwMSIsInVzZXJSb2xlcyI6Ilt7XCJpZFwiOjEsXCJ1c2VyX2lkXCI6MSxcInVzZXJfbmFtZVwiOlwibHNqMDAxXCIsXCJ2YWx1ZVwiOlwiXCJ9XSJ9.2wNr4tF9xRNMBxqdsPGzudllFjp9LID7-o4DlKhPXlI'),(3,'admin','2020-08-03 10:47:10','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTY0NTE2MjksIm9yaWdfaWF0IjoxNTk2NDQ4MDI5LCJ1c2VyTmFtZSI6ImFkbWluIiwidXNlclJvbGVzIjoiW3tcImlkXCI6MixcInVzZXJfaWRcIjoyLFwidXNlcl9uYW1lXCI6XCJhZG1pblwiLFwidmFsdWVcIjpcImFkbWluXCJ9XSJ9.EjX6PuTCSYzt205km_7bNTkLJyy6JwqYQAVf04BhpkE'),(4,'lsj001','2020-08-03 10:53:17','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTY0NTE5OTYsIm9yaWdfaWF0IjoxNTk2NDQ4Mzk2LCJ1c2VyTmFtZSI6ImxzajAwMSIsInVzZXJSb2xlcyI6Ilt7XCJpZFwiOjEsXCJ1c2VyX2lkXCI6MSxcInVzZXJfbmFtZVwiOlwibHNqMDAxXCIsXCJ2YWx1ZVwiOlwiXCJ9XSJ9.p-rNfxD5VuhK5iXArgnHGKgW6JygJ0J4AFEbjlz9SmU'),(5,'lsj001','2020-08-03 11:08:07','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTY0NTI4ODcsIm9yaWdfaWF0IjoxNTk2NDQ5Mjg3LCJ1c2VyTmFtZSI6ImxzajAwMSIsInVzZXJSb2xlcyI6Ilt7XCJpZFwiOjEsXCJ1c2VyX2lkXCI6MSxcInVzZXJfbmFtZVwiOlwibHNqMDAxXCIsXCJ2YWx1ZVwiOlwiXCJ9XSJ9.Db7BEx4S-VRaOkaToBzlRKAmb7XHYCuQrAzzoBCrKds'),(6,'lsj001','2020-08-04 17:01:55','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTY1NjA1MTQsIm9yaWdfaWF0IjoxNTk2NTU2OTE0LCJ1c2VyTmFtZSI6ImxzajAwMSIsInVzZXJSb2xlcyI6Ilt7XCJpZFwiOjEsXCJ1c2VyX2lkXCI6MSxcInVzZXJfbmFtZVwiOlwibHNqMDAxXCIsXCJ2YWx1ZVwiOlwiXCJ9XSJ9.pipuCTsW7qwIKV2tJt4vO0mAal1-B5ULTHCIw2EqzrQ'),(7,'admin','2020-08-04 17:03:12','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTY1NjA1OTEsIm9yaWdfaWF0IjoxNTk2NTU2OTkxLCJ1c2VyTmFtZSI6ImFkbWluIiwidXNlclJvbGVzIjoiW3tcImlkXCI6MixcInVzZXJfaWRcIjoyLFwidXNlcl9uYW1lXCI6XCJhZG1pblwiLFwidmFsdWVcIjpcImFkbWluXCJ9XSJ9.ujy9Gu9-drHKnmbJA_fw_ZdMYaIelV1kqqNM1Swq6p4');
/*!40000 ALTER TABLE `tokens` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `mobile` varchar(255) DEFAULT NULL,
  `username` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `gender` int DEFAULT NULL,
  `avatar` varchar(255) DEFAULT NULL,
  `label` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `user_type` int DEFAULT NULL,
  `deleted` int DEFAULT NULL,
  `state` int DEFAULT NULL,
  `extend` varchar(255) DEFAULT NULL,
  `contact_person` varchar(255) DEFAULT NULL,
  `introductory` text,
  `province` varchar(255) DEFAULT NULL,
  `city` varchar(255) DEFAULT NULL,
  `county` varchar(255) DEFAULT NULL,
  `street` varchar(255) DEFAULT NULL,
  `address` varchar(255) DEFAULT NULL,
  `branches_name` varchar(255) DEFAULT NULL,
  `legal_person` varchar(255) DEFAULT NULL,
  `legal_identity_card` varchar(255) DEFAULT NULL,
  `created_by` varchar(255) DEFAULT NULL,
  `modified_by` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'2020-08-03 09:22:56','2020-08-03 09:22:56','13702290109','lsj001','654321',1,'https://zbj-bucket1.oss-cn-shenzhen.aliyuncs.com/avatar.JPG','','',1,0,1,'','李示佳','','','','','','','','','','',''),(2,'2020-08-03 09:44:34','2020-08-03 09:44:34','13702290109','admin','654321',1,'https://zbj-bucket1.oss-cn-shenzhen.aliyuncs.com/avatar.JPG','','',3,0,1,'','管理员','','','','','','','','','','','');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2020-08-05 23:40:22
