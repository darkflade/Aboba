<?php
require_once __DIR__ . '/../../../private/vendor/autoload.php';

use PhpOffice\PhpWord\PhpWord;
use PhpOffice\PhpWord\IOFactory;

$allowedOrigins = [
    'https://evil.net',
    'https://server.go'
];

$origin = isset($_SERVER['HTTP_ORIGIN']) ? $_SERVER['HTTP_ORIGIN'] : '';

if (in_array($origin, $allowedOrigins)) {
    header("Access-Control-Allow-Origin: $origin");
} else {
    http_response_code(403);
    echo json_encode(['error' => 'CORS policy: Origin not allowed']);
    exit;
}

header("Access-Control-Allow-Methods: GET, OPTIONS");
header("Access-Control-Allow-Headers: Content-Type");

// Получаем id отдела
$deptId = isset($_GET['id']) ? (int)$_GET['id'] : 0;
if ($deptId < 1) {
    http_response_code(400);
    echo "Некорректный ID отдела";
    exit;
}


$db = new PDO(
    'mysql:host=127.127.126.1;dbname=WorkDB;charset=utf8mb4',
    'root',
    '',
    [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
);

// Запрос для получения информации об отделе
$stmt = $db->prepare("
        SELECT ОТД_СОТР_ЗАРП AS expectedSalary,
               ОТД_РАЗМ       AS expectedCount,
               ОТД_РУК        AS bossId
        FROM ОТДЕЛЫ
        WHERE ОТД_НОМЕР = :dept
        ");
$stmt->execute([':dept' => $deptId]);
$deptInfo = $stmt->fetch(PDO::FETCH_ASSOC);
if (!$deptInfo) {
    return $response->withStatus(404)->write("Отдел №{$deptId} не найден");
}
$expectedSalary = (float)$deptInfo['expectedSalary'];
$expectedCount  = (int)$deptInfo['expectedCount'];
$bossId         = (int)$deptInfo['bossId'];


// Запрос для получения сотрудников отдела
$stmt = $db->prepare("
    SELECT СЛУ_НОМЕР AS id, СЛУ_ИМЯ AS name, СЛУ_СТАТ AS status, СЛУ_ЗАРП AS salary
    FROM СЛУЖАЩИЕ
    WHERE СЛУ_ОТД_НОМЕР = :dept
");
$stmt->execute([':dept' => $deptId]);
$employees = [];
    $actualSalary = 0.0;
    $actualCount  = 0;
    while ($row = $stmt->fetch(PDO::FETCH_ASSOC)) {
        $row['salary'] = (float)$row['salary'];
        $employees[]   = $row;
        $actualSalary += $row['salary'];
        $actualCount++;
    }

    if ($actualCount === 0) {
        http_response_code(404);
        echo "Нет сотрудников в отделе №{$deptId}";
        exit;
    }

// Генерируем Word-документ
$phpWord = new PhpWord();

// Стили
$phpWord->addTitleStyle(1, ['bold'=>true, 'size'=>18, 'color'=>'2E74B5']);
$phpWord->addTitleStyle(2, ['bold'=>true, 'size'=>14, 'color'=>'4F81BD']);
$tableStyle = [
    'borderSize' => 6,
    'borderColor'=> '4F81BD',
    'cellMargin' => 80,
    'alignment'  => \PhpOffice\PhpWord\SimpleType\JcTable::CENTER,
];
$phpWord->addTableStyle("DeptTable", $tableStyle, [
    'borderBottomColor'=>'4F81BD',
    'borderBottomSize'=>12
]);
$headerFont = ['bold'=>true, 'color'=>'FFFFFF', 'size'=>12];
$headerCellStyle = ['bgColor'=>'4F81BD'];
$bossCellStyle = ['bgColor'=>'E6F0FA']; //E6F5E1
$rowAltStyle = ['bgColor'=>'E7E6E6'];

// Раздел
$section = $phpWord->addSection([
    'marginTop'=>800, 'marginBottom'=>800, 'marginLeft'=>1200, 'marginRight'=>1200
]);

// Заголовок документа
$section->addTitle("Сотрудники отдела №{$deptId}", 1);

// Таблица
$table = $section->addTable("DeptTable");

// Шапка
$table->addRow(500);
foreach (['ID','Имя','Статус','Зарплата'] as $txt) {
    $table->addCell(2000, $headerCellStyle)->addText($txt, $headerFont, ['alignment'=>'center']);
}

// Данные (выделяем руководителя жёлтым фоном, чередуем строки)
$rowIdx = 0;
foreach ($employees as $e) {
    $table->addRow(400);
    $isBoss = $e['id'] === $bossId;
    $cellStyle = $isBoss ? $bossCellStyle : ($rowIdx % 2 ? $rowAltStyle : []);
    $font = $isBoss ? ['bold'=>true, 'color'=>'1565C0'] : []; //4CAF50
    $table->addCell(1000, $cellStyle)->addText($e['id'], $font);
    $table->addCell(4000, $cellStyle)->addText($e['name'], $font);
    $table->addCell(2000, $cellStyle)->addText($e['status'], $font);
    $table->addCell(2000, $cellStyle)->addText(number_format($e['salary'], 2, ',', ' '), $font);
    $rowIdx++;
}

// Статистика
$section->addTextBreak(2);
$statsFont = ['italic'=>true, 'color'=>'888888', 'size'=>11];
$section->addTitle("Статистика по отделу", 2);

$section->addText("Ожидаемые (по таблице ОТДЕЛЫ):", $statsFont);
$section->addListItem("Суммарная зарплата: " . number_format($expectedSalary, 2, ',', ' '), 0, ['color'=>'2E74B5']);
$section->addListItem("Количество сотрудников: " . $expectedCount, 0, ['color'=>'2E74B5']);

$section->addText("Фактические (по выборке СЛУЖАЩИЕ):", $statsFont);
$section->addListItem("Суммарная зарплата: " . number_format($actualSalary, 2, ',', ' '), 0, ['color'=>'C00000']);
$section->addListItem("Количество сотрудников: " . $actualCount, 0, ['color'=>'C00000']);


// Сохраняем в память и отдаём клиенту
$writer = IOFactory::createWriter($phpWord, 'Word2007');
ob_start();
$writer->save('php://output');
$docxContent = ob_get_clean();

$filename = "department_{$deptId}_employees.docx";

// Установка заголовков для скачивания файла
header('Content-Description: File Transfer');
header('Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document');
header('Content-Disposition: attachment; filename="' . $filename . '"');
header('Content-Length: ' . strlen($docxContent));
echo $docxContent;
exit;
