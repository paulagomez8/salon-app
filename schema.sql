-- ============================================================
-- salon_db — Schema completo
-- Última actualización: abril 2026
-- ============================================================

CREATE DATABASE IF NOT EXISTS salon_db
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE salon_db;

-- ── Admin ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `admin_config` (
  `id`       int(11)      NOT NULL AUTO_INCREMENT,
  `usuario`  varchar(100) NOT NULL,
  `password` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `admin_config` (`usuario`, `password`) VALUES ('clemence', 'salon2026')
  ON DUPLICATE KEY UPDATE `usuario` = `usuario`;

-- ── Empleadas ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `empleadas` (
  `id`     int(11)      NOT NULL AUTO_INCREMENT,
  `nombre` varchar(100) NOT NULL,
  `activa` tinyint(1)   DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `empleadas` (`id`, `nombre`, `activa`) VALUES (1, 'Clémence', 1)
  ON DUPLICATE KEY UPDATE `nombre` = `nombre`;

-- ── Clientes ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `clientes` (
  `id`         int(11)      NOT NULL AUTO_INCREMENT,
  `nombre`     varchar(100) NOT NULL,
  `telefono`   varchar(20)  NOT NULL,
  `email`      varchar(100) DEFAULT NULL,
  `creado_en`  timestamp    NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `telefono` (`telefono`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ── Servicios ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `servicios` (
  `id`                     int(11)        NOT NULL AUTO_INCREMENT,
  `nombre`                 varchar(100)   NOT NULL,
  `descripcion`            text           DEFAULT NULL,
  `duracion_minutos`       int(11)        NOT NULL,
  `duracion_activa_minutos` int(11)       NOT NULL,
  `permite_paralelo`       tinyint(1)     DEFAULT 0,
  `activo`                 tinyint(1)     DEFAULT 1,
  `precio`                 decimal(10,2)  NOT NULL DEFAULT 0.00,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ── Citas ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `citas` (
  `id`                     int(11)     NOT NULL AUTO_INCREMENT,
  `cliente_id`             int(11)     NOT NULL,
  `empleada_id`            int(11)     NOT NULL,
  `servicio_id`            int(11)     NOT NULL,
  `fecha_hora`             datetime    NOT NULL,
  `duracion_minutos`       int(11)     NOT NULL DEFAULT 0,
  `duracion_activa_minutos` int(11)    NOT NULL DEFAULT 0,
  `permite_paralelo`       tinyint(1)  DEFAULT 0,
  `notas`                  text        DEFAULT NULL,
  `creado_en`              timestamp   NULL DEFAULT current_timestamp(),
  `estado`                 varchar(20) NOT NULL DEFAULT 'activa',
  PRIMARY KEY (`id`),
  KEY `cliente_id`  (`cliente_id`),
  KEY `empleada_id` (`empleada_id`),
  KEY `servicio_id` (`servicio_id`),
  CONSTRAINT `citas_ibfk_1` FOREIGN KEY (`cliente_id`)  REFERENCES `clientes`  (`id`),
  CONSTRAINT `citas_ibfk_2` FOREIGN KEY (`empleada_id`) REFERENCES `empleadas` (`id`),
  CONSTRAINT `citas_ibfk_3` FOREIGN KEY (`servicio_id`) REFERENCES `servicios` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ── Bloqueos ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `bloqueos` (
  `id`          int(11)      NOT NULL AUTO_INCREMENT,
  `empleada_id` int(11)      NOT NULL,
  `titulo`      varchar(100) DEFAULT NULL,
  `fecha`       date         DEFAULT NULL,
  `hora_inicio` time         NOT NULL,
  `hora_fin`    time         NOT NULL,
  `dia_semana`  int(11)      DEFAULT NULL,
  `recurrente`  tinyint(1)   DEFAULT 0,
  `activo`      tinyint(1)   NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `empleada_id` (`empleada_id`),
  CONSTRAINT `bloqueos_ibfk_1` FOREIGN KEY (`empleada_id`) REFERENCES `empleadas` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
