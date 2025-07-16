-- Схема базы данных для приложения "Отделы - Служащие" (MySQL)
-- Без директив DELIMITER, готово к выполнению через MySQL-клиент или Exec из Go (MultiStatements=true)

-- Создание таблицы ОТДЕЛЫ
CREATE TABLE `ОТДЕЛЫ` (
  `ОТД_НОМЕР` INT AUTO_INCREMENT PRIMARY KEY,
  `ОТД_РУК` VARCHAR(100) NOT NULL,
  `ОТД_СОТР_ЗАРП` DECIMAL(12,2) NOT NULL DEFAULT 0,
  `ОТД_РАЗМ` INT NOT NULL CHECK (`ОТД_РАЗМ` > 0)
) ENGINE=InnoDB;

-- Создание таблицы СЛУЖАЩИЕ
CREATE TABLE `СЛУЖАЩИЕ` (
  `СЛУ_НОМЕР` INT AUTO_INCREMENT PRIMARY KEY,
  `СЛУ_ИМЯ` VARCHAR(100) NOT NULL,
  `СЛУ_СТАТ` VARCHAR(50) NOT NULL,
  `СЛУ_ЗАРП` DECIMAL(12,2) NOT NULL CHECK (`СЛУ_ЗАРП` >= 0),
  `СЛУ_ОТД_НОМЕР` INT NOT NULL,
  INDEX `idx_служ_отд` (`СЛУ_ОТД_НОМЕР`),
  CONSTRAINT `fk_служ_отд` FOREIGN KEY (`СЛУ_ОТД_НОМЕР`) REFERENCES `ОТДЕЛЫ`(`ОТД_НОМЕР`) ON DELETE RESTRICT
) ENGINE=InnoDB;

-- Триггер: после вставки сотрудника
CREATE TRIGGER `after_insert_служ`
AFTER INSERT ON `СЛУЖАЩИЕ`
FOR EACH ROW
BEGIN
  UPDATE `ОТДЕЛЫ`
  SET `ОТД_СОТР_ЗАРП` = `ОТД_СОТР_ЗАРП` + NEW.`СЛУ_ЗАРП`
  WHERE `ОТД_НОМЕР` = NEW.`СЛУ_ОТД_НОМЕР`;
END;

-- Триггер: после удаления сотрудника
CREATE TRIGGER `after_delete_служ`
AFTER DELETE ON `СЛУЖАЩИЕ`
FOR EACH ROW
BEGIN
  UPDATE `ОТДЕЛЫ`
  SET `ОТД_СОТР_ЗАРП` = `ОТД_СОТР_ЗАРП` - OLD.`СЛУ_ЗАРП`
  WHERE `ОТД_НОМЕР` = OLD.`СЛУ_ОТД_НОМЕР`;
END;

-- Триггер: после обновления сотрудника
CREATE TRIGGER `after_update_служ`
AFTER UPDATE ON `СЛУЖАЩИЕ`
FOR EACH ROW
BEGIN
  -- корректируем сумму при смене зарплаты
  IF NEW.`СЛУ_ЗАРП` <> OLD.`СЛУ_ЗАРП` THEN
    UPDATE `ОТДЕЛЫ`
    SET `ОТД_СОТР_ЗАРП` = `ОТД_СОТР_ЗАРП` + (NEW.`СЛУ_ЗАРП` - OLD.`СЛУ_ЗАРП`)
    WHERE `ОТД_НОМЕР` = NEW.`СЛУ_ОТД_НОМЕР`;
  END IF;
  -- корректируем при смене отдела
  IF NEW.`СЛУ_ОТД_НОМЕР` <> OLD.`СЛУ_ОТД_НОМЕР` THEN
    UPDATE `ОТДЕЛЫ`
    SET `ОТД_СОТР_ЗАРП` = `ОТД_СОТР_ЗАРП` - OLD.`СЛУ_ЗАРП`
    WHERE `ОТД_НОМЕР` = OLD.`СЛУ_ОТД_НОМЕР`;
    UPDATE `ОТДЕЛЫ`
    SET `ОТД_СОТР_ЗАРП` = `ОТД_СОТР_ЗАРП` + NEW.`СЛУ_ЗАРП`
    WHERE `ОТД_НОМЕР` = NEW.`СЛУ_ОТД_НОМЕР`;
  END IF;
END;

-- Примеры транзакций (будут использоваться также в Go):
-- добавление нового сотрудника
START TRANSACTION;
  INSERT INTO `СЛУЖАЩИЕ` (`СЛУ_ИМЯ`,`СЛУ_СТАТ`,`СЛУ_ЗАРП`,`СЛУ_ОТД_НОМЕР`)
  VALUES ('Иван Иванов','менеджер',50000.00,1);
COMMIT;

-- удаление сотрудника
START TRANSACTION;
  DELETE FROM `СЛУЖАЩИЕ` WHERE `СЛУ_НОМЕР`=42;
COMMIT;

-- изменение зарплаты сотрудника
START TRANSACTION;
  UPDATE `СЛУЖАЩИЕ` SET `СЛУ_ЗАРП`=55000.00 WHERE `СЛУ_НОМЕР`=42;
COMMIT;

-- смена руководителя отдела
START TRANSACTION;
  UPDATE `ОТДЕЛЫ` SET `ОТД_РУК`='Петр Петров' WHERE `ОТД_НОМЕР`=1;
COMMIT;
