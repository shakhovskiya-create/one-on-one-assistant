// EKF Hub UI Kit - Figma Plugin
// Generates UI components in EKF Hub design style

// =====================
// COLOR PALETTE
// =====================
var COLORS = {
  // Brand
  ekfRed: { r: 229/255, g: 57/255, b: 53/255 },      // #E53935
  ekfDark: { r: 26/255, g: 26/255, b: 46/255 },      // #1a1a2e

  // Grays
  white: { r: 1, g: 1, b: 1 },
  gray50: { r: 249/255, g: 250/255, b: 251/255 },    // #f9fafb
  gray100: { r: 243/255, g: 244/255, b: 246/255 },   // #f3f4f6
  gray200: { r: 229/255, g: 231/255, b: 235/255 },   // #e5e7eb
  gray300: { r: 209/255, g: 213/255, b: 219/255 },   // #d1d5db
  gray400: { r: 156/255, g: 163/255, b: 175/255 },   // #9ca3af
  gray500: { r: 107/255, g: 114/255, b: 128/255 },   // #6b7280
  gray600: { r: 75/255, g: 85/255, b: 99/255 },      // #4b5563
  gray700: { r: 55/255, g: 65/255, b: 81/255 },      // #374151
  gray800: { r: 31/255, g: 41/255, b: 55/255 },      // #1f2937
  gray900: { r: 17/255, g: 24/255, b: 39/255 },      // #111827

  // Status colors
  blue50: { r: 239/255, g: 246/255, b: 255/255 },
  blue100: { r: 219/255, g: 234/255, b: 254/255 },
  blue500: { r: 59/255, g: 130/255, b: 246/255 },
  blue600: { r: 37/255, g: 99/255, b: 235/255 },

  green50: { r: 240/255, g: 253/255, b: 244/255 },
  green100: { r: 220/255, g: 252/255, b: 231/255 },
  green500: { r: 34/255, g: 197/255, b: 94/255 },

  yellow50: { r: 254/255, g: 252/255, b: 232/255 },
  yellow100: { r: 254/255, g: 249/255, b: 195/255 },
  yellow400: { r: 250/255, g: 204/255, b: 21/255 },

  orange50: { r: 255/255, g: 247/255, b: 237/255 },
  orange100: { r: 255/255, g: 237/255, b: 213/255 },
  orange400: { r: 251/255, g: 146/255, b: 60/255 },
  orange600: { r: 234/255, g: 88/255, b: 12/255 },

  purple50: { r: 250/255, g: 245/255, b: 255/255 },
  purple100: { r: 243/255, g: 232/255, b: 255/255 },
  purple400: { r: 192/255, g: 132/255, b: 252/255 },
  purple500: { r: 168/255, g: 85/255, b: 247/255 },

  red50: { r: 254/255, g: 242/255, b: 242/255 },
  red100: { r: 254/255, g: 226/255, b: 226/255 },
  red500: { r: 239/255, g: 68/255, b: 68/255 },
  red600: { r: 220/255, g: 38/255, b: 38/255 },
  red700: { r: 185/255, g: 28/255, b: 28/255 },
};

// =====================
// TYPOGRAPHY
// =====================
function loadFonts() {
  return Promise.all([
    figma.loadFontAsync({ family: "Inter", style: "Regular" }),
    figma.loadFontAsync({ family: "Inter", style: "Medium" }),
    figma.loadFontAsync({ family: "Inter", style: "SemiBold" }),
    figma.loadFontAsync({ family: "Inter", style: "Bold" })
  ]);
}

// =====================
// HELPER FUNCTIONS
// =====================
function createFrame(name, width, height) {
  var frame = figma.createFrame();
  frame.name = name;
  frame.resize(width, height);
  frame.fills = [];
  return frame;
}

function setFill(node, color, opacity) {
  if (opacity === undefined) opacity = 1;
  node.fills = [{ type: 'SOLID', color: color, opacity: opacity }];
}

function setStroke(node, color, weight) {
  if (weight === undefined) weight = 1;
  node.strokes = [{ type: 'SOLID', color: color }];
  if ('strokeWeight' in node) {
    node.strokeWeight = weight;
  }
}

function createText(content, fontSize, fontWeight, color) {
  if (fontWeight === undefined) fontWeight = "Regular";
  if (color === undefined) color = COLORS.gray900;
  var text = figma.createText();
  text.characters = content;
  text.fontSize = fontSize;
  text.fontName = { family: "Inter", style: fontWeight };
  text.fills = [{ type: 'SOLID', color: color }];
  return text;
}

function createRoundedRect(name, width, height, radius, color) {
  var rect = figma.createRectangle();
  rect.name = name;
  rect.resize(width, height);
  rect.cornerRadius = radius;
  setFill(rect, color);
  return rect;
}

function createAvatar(name, initial, size, bgColor) {
  var frame = createFrame(name, size, size);
  frame.cornerRadius = size / 2;
  setFill(frame, bgColor);

  var text = createText(initial, size * 0.4, "Medium", COLORS.white);
  text.x = (size - text.width) / 2;
  text.y = (size - text.height) / 2;
  frame.appendChild(text);

  return frame;
}

// =====================
// UI COMPONENTS
// =====================

// Sidebar
function createSidebar() {
  var sidebar = createFrame("Sidebar", 256, 800);
  setFill(sidebar, COLORS.ekfDark);
  sidebar.layoutMode = "VERTICAL";
  sidebar.paddingTop = 0;
  sidebar.paddingBottom = 0;
  sidebar.itemSpacing = 0;

  // Logo section
  var logoSection = createFrame("Logo Section", 256, 64);
  setFill(logoSection, COLORS.ekfDark);
  logoSection.layoutMode = "HORIZONTAL";
  logoSection.paddingLeft = 16;
  logoSection.paddingRight = 16;
  logoSection.paddingTop = 16;
  logoSection.paddingBottom = 16;
  logoSection.itemSpacing = 12;
  logoSection.primaryAxisAlignItems = "MIN";
  logoSection.counterAxisAlignItems = "CENTER";

  // EKF Logo box
  var logoBox = createRoundedRect("Logo Box", 48, 32, 4, COLORS.ekfRed);
  var logoText = createText("EKF", 16, "Bold", COLORS.white);

  var logoFrame = createFrame("Logo", 48, 32);
  logoFrame.appendChild(logoBox);
  logoBox.x = 0;
  logoBox.y = 0;
  logoText.x = (48 - logoText.width) / 2;
  logoText.y = (32 - logoText.height) / 2;
  logoFrame.appendChild(logoText);

  var hubText = createText("Hub", 18, "SemiBold", COLORS.white);

  logoSection.appendChild(logoFrame);
  logoSection.appendChild(hubText);

  // Divider
  var divider = createRoundedRect("Divider", 256, 1, 0, COLORS.gray700);

  // Navigation
  var nav = createFrame("Navigation", 256, 600);
  setFill(nav, COLORS.ekfDark);
  nav.layoutMode = "VERTICAL";
  nav.paddingLeft = 16;
  nav.paddingRight = 16;
  nav.paddingTop = 16;
  nav.itemSpacing = 4;

  var menuItems = [
    { icon: "dashboard", label: "Дашборд", active: false },
    { icon: "users", label: "Сотрудники", active: false },
    { icon: "tasks", label: "Задачи", active: true },
    { icon: "calendar", label: "Встречи", active: false },
    { icon: "support", label: "Service Desk", active: false },
    { icon: "messenger", label: "Сообщения", active: false },
    { icon: "mail", label: "Почта", active: false },
    { icon: "analytics", label: "Аналитика", active: false },
  ];

  menuItems.forEach(function(item) {
    var menuItem = createFrame("Menu Item - " + item.label, 224, 44);
    menuItem.cornerRadius = 8;
    menuItem.layoutMode = "HORIZONTAL";
    menuItem.paddingLeft = 16;
    menuItem.paddingRight = 16;
    menuItem.itemSpacing = 12;
    menuItem.counterAxisAlignItems = "CENTER";

    if (item.active) {
      setFill(menuItem, COLORS.ekfRed);
    } else {
      menuItem.fills = [];
    }

    // Icon placeholder
    var iconPlaceholder = createRoundedRect("Icon", 20, 20, 2, item.active ? COLORS.white : COLORS.gray300);
    iconPlaceholder.opacity = 0.5;

    var labelText = createText(item.label, 14, "Regular", item.active ? COLORS.white : COLORS.gray300);

    menuItem.appendChild(iconPlaceholder);
    menuItem.appendChild(labelText);
    nav.appendChild(menuItem);
  });

  sidebar.appendChild(logoSection);
  sidebar.appendChild(divider);
  sidebar.appendChild(nav);

  return sidebar;
}

// Header
function createHeader(title) {
  if (title === undefined) title = "Задачи";
  var header = createFrame("Header", 1200, 56);
  setFill(header, COLORS.white);
  setStroke(header, COLORS.gray200);
  header.layoutMode = "HORIZONTAL";
  header.paddingLeft = 16;
  header.paddingRight = 16;
  header.primaryAxisAlignItems = "SPACE_BETWEEN";
  header.counterAxisAlignItems = "CENTER";

  var titleText = createText(title, 20, "SemiBold", COLORS.gray800);

  var userSection = createFrame("User Section", 200, 40);
  userSection.layoutMode = "HORIZONTAL";
  userSection.itemSpacing = 8;
  userSection.counterAxisAlignItems = "CENTER";
  userSection.layoutAlign = "STRETCH";
  userSection.primaryAxisSizingMode = "AUTO";

  var userName = createText("Антон Саховский", 14, "Medium", COLORS.gray700);
  var avatar = createAvatar("User Avatar", "А", 32, COLORS.ekfRed);

  userSection.appendChild(userName);
  userSection.appendChild(avatar);

  header.appendChild(titleText);
  header.appendChild(userSection);

  return header;
}

// Button
function createButton(label, variant, size) {
  if (variant === undefined) variant = "primary";
  if (size === undefined) size = "md";
  var height = size === "sm" ? 32 : 40;
  var fontSize = size === "sm" ? 13 : 14;
  var paddingX = size === "sm" ? 12 : 16;

  var button = createFrame("Button - " + label, 100, height);
  button.cornerRadius = 8;
  button.layoutMode = "HORIZONTAL";
  button.paddingLeft = paddingX;
  button.paddingRight = paddingX;
  button.primaryAxisSizingMode = "AUTO";
  button.counterAxisAlignItems = "CENTER";
  button.primaryAxisAlignItems = "CENTER";
  button.itemSpacing = 6;

  if (variant === "primary") {
    setFill(button, COLORS.ekfRed);
  } else {
    setFill(button, COLORS.white);
    setStroke(button, COLORS.gray200);
  }

  var text = createText(label, fontSize, "Medium", variant === "primary" ? COLORS.white : COLORS.gray600);
  button.appendChild(text);

  return button;
}

// Input Field
function createInput(placeholder, width) {
  if (placeholder === undefined) placeholder = "Поиск...";
  if (width === undefined) width = 200;
  var input = createFrame("Input", width, 36);
  input.cornerRadius = 8;
  setFill(input, COLORS.white);
  setStroke(input, COLORS.gray200);
  input.layoutMode = "HORIZONTAL";
  input.paddingLeft = 12;
  input.paddingRight = 12;
  input.counterAxisAlignItems = "CENTER";

  var text = createText(placeholder, 14, "Regular", COLORS.gray400);
  input.appendChild(text);

  return input;
}

// Select/Dropdown
function createSelect(label, width) {
  if (label === undefined) label = "Все проекты";
  if (width === undefined) width = 150;
  var select = createFrame("Select", width, 36);
  select.cornerRadius = 8;
  setFill(select, COLORS.white);
  setStroke(select, COLORS.gray200);
  select.layoutMode = "HORIZONTAL";
  select.paddingLeft = 12;
  select.paddingRight = 12;
  select.counterAxisAlignItems = "CENTER";
  select.primaryAxisAlignItems = "SPACE_BETWEEN";

  var text = createText(label, 14, "Regular", COLORS.gray700);
  var arrow = createText("▼", 10, "Regular", COLORS.gray400);

  select.appendChild(text);
  select.appendChild(arrow);

  return select;
}

// Task Card
function createTaskCard(taskId, title, priority, storyPoints, assigneeInitial) {
  if (priority === undefined) priority = "medium";
  var card = createFrame("Task Card - " + taskId, 260, 120);
  card.cornerRadius = 8;
  setFill(card, COLORS.white);
  setStroke(card, COLORS.gray200);
  card.layoutMode = "VERTICAL";
  card.paddingLeft = 12;
  card.paddingRight = 12;
  card.paddingTop = 12;
  card.paddingBottom = 12;
  card.itemSpacing = 8;

  // Priority indicator
  var priorityColors = {
    critical: COLORS.red500,
    high: COLORS.orange600,
    medium: COLORS.yellow400,
    low: COLORS.blue500
  };

  // Add left border for priority
  card.strokeLeftWeight = 3;
  card.strokes = [{ type: 'SOLID', color: priorityColors[priority] }];

  // Header row
  var headerRow = createFrame("Header Row", 236, 20);
  headerRow.layoutMode = "HORIZONTAL";
  headerRow.primaryAxisAlignItems = "SPACE_BETWEEN";
  headerRow.counterAxisAlignItems = "CENTER";
  headerRow.fills = [];

  var idText = createText(taskId, 11, "Regular", COLORS.gray400);

  if (storyPoints) {
    var spBadge = createFrame("SP Badge", 40, 18);
    spBadge.cornerRadius = 4;
    setFill(spBadge, COLORS.gray100);
    spBadge.layoutMode = "HORIZONTAL";
    spBadge.primaryAxisAlignItems = "CENTER";
    spBadge.counterAxisAlignItems = "CENTER";
    var spText = createText(storyPoints + " SP", 10, "Medium", COLORS.gray600);
    spBadge.appendChild(spText);
    headerRow.appendChild(idText);
    headerRow.appendChild(spBadge);
  } else {
    headerRow.appendChild(idText);
  }

  // Title
  var titleText = createText(title, 13, "Medium", COLORS.gray900);
  titleText.layoutAlign = "STRETCH";
  titleText.textAutoResize = "HEIGHT";

  // Footer row
  var footerRow = createFrame("Footer Row", 236, 24);
  footerRow.layoutMode = "HORIZONTAL";
  footerRow.primaryAxisAlignItems = "SPACE_BETWEEN";
  footerRow.counterAxisAlignItems = "CENTER";
  footerRow.fills = [];

  // Tags area
  var tagsArea = createFrame("Tags", 100, 20);
  tagsArea.fills = [];

  // Assignee
  if (assigneeInitial) {
    var avatar = createAvatar("Assignee", assigneeInitial, 24, COLORS.ekfRed);
    footerRow.appendChild(tagsArea);
    footerRow.appendChild(avatar);
  }

  card.appendChild(headerRow);
  card.appendChild(titleText);
  card.appendChild(footerRow);

  return card;
}

// Kanban Column
function createKanbanColumn(name, count, storyPoints, bgColor, dotColor) {
  var column = createFrame("Column - " + name, 280, 600);
  column.cornerRadius = 8;
  setFill(column, bgColor);
  column.layoutMode = "VERTICAL";
  column.itemSpacing = 0;

  // Header
  var header = createFrame("Column Header", 280, 48);
  header.layoutMode = "HORIZONTAL";
  header.paddingLeft = 12;
  header.paddingRight = 12;
  header.primaryAxisAlignItems = "SPACE_BETWEEN";
  header.counterAxisAlignItems = "CENTER";
  header.fills = [];

  var leftSection = createFrame("Left", 150, 24);
  leftSection.layoutMode = "HORIZONTAL";
  leftSection.itemSpacing = 8;
  leftSection.counterAxisAlignItems = "CENTER";
  leftSection.fills = [];

  var dot = createRoundedRect("Dot", 8, 8, 4, dotColor);
  var nameText = createText(name, 13, "Medium", COLORS.gray800);

  var countBadge = createFrame("Count", 28, 20);
  countBadge.cornerRadius = 10;
  setFill(countBadge, COLORS.gray200);
  countBadge.layoutMode = "HORIZONTAL";
  countBadge.primaryAxisAlignItems = "CENTER";
  countBadge.counterAxisAlignItems = "CENTER";
  var countText = createText(count.toString(), 11, "Medium", COLORS.gray600);
  countBadge.appendChild(countText);

  leftSection.appendChild(dot);
  leftSection.appendChild(nameText);
  leftSection.appendChild(countBadge);

  var spText = createText(storyPoints + " SP", 11, "Regular", COLORS.gray500);

  header.appendChild(leftSection);
  header.appendChild(spText);

  // Cards container
  var cardsContainer = createFrame("Cards", 280, 540);
  cardsContainer.layoutMode = "VERTICAL";
  cardsContainer.paddingLeft = 8;
  cardsContainer.paddingRight = 8;
  cardsContainer.paddingTop = 8;
  cardsContainer.itemSpacing = 8;
  cardsContainer.fills = [];

  column.appendChild(header);
  column.appendChild(cardsContainer);

  return column;
}

// Filter Bar
function createFilterBar() {
  var filterBar = createFrame("Filter Bar", 1150, 56);
  filterBar.cornerRadius = 8;
  setFill(filterBar, COLORS.white);
  filterBar.layoutMode = "HORIZONTAL";
  filterBar.paddingLeft = 12;
  filterBar.paddingRight = 12;
  filterBar.itemSpacing = 12;
  filterBar.counterAxisAlignItems = "CENTER";

  // Search
  var search = createInput("Поиск задач...", 200);

  // Filters
  var projectFilter = createSelect("Все проекты", 140);
  var assigneeFilter = createSelect("Все исполнители", 150);
  var sprintFilter = createSelect("Все спринты", 140);
  var releaseFilter = createSelect("Все релизы", 120);

  // Spacer
  var spacer = createFrame("Spacer", 200, 1);
  spacer.layoutGrow = 1;
  spacer.fills = [];

  // View toggle
  var viewToggle = createFrame("View Toggle", 160, 36);
  viewToggle.cornerRadius = 8;
  setStroke(viewToggle, COLORS.gray200);
  viewToggle.layoutMode = "HORIZONTAL";
  viewToggle.fills = [];

  var listBtn = createFrame("List Btn", 80, 36);
  setFill(listBtn, COLORS.white);
  listBtn.layoutMode = "HORIZONTAL";
  listBtn.primaryAxisAlignItems = "CENTER";
  listBtn.counterAxisAlignItems = "CENTER";
  var listText = createText("Список", 13, "Regular", COLORS.gray600);
  listBtn.appendChild(listText);

  var kanbanBtn = createFrame("Kanban Btn", 80, 36);
  setFill(kanbanBtn, COLORS.ekfRed);
  kanbanBtn.layoutMode = "HORIZONTAL";
  kanbanBtn.primaryAxisAlignItems = "CENTER";
  kanbanBtn.counterAxisAlignItems = "CENTER";
  var kanbanText = createText("Kanban", 13, "Medium", COLORS.white);
  kanbanBtn.appendChild(kanbanText);

  viewToggle.appendChild(listBtn);
  viewToggle.appendChild(kanbanBtn);

  filterBar.appendChild(search);
  filterBar.appendChild(projectFilter);
  filterBar.appendChild(assigneeFilter);
  filterBar.appendChild(sprintFilter);
  filterBar.appendChild(releaseFilter);
  filterBar.appendChild(spacer);
  filterBar.appendChild(viewToggle);

  return filterBar;
}

// =====================
// PAGE GENERATORS
// =====================

function generateTasksPage() {
  var page = createFrame("Tasks Page", 1440, 900);
  setFill(page, COLORS.gray100);
  page.layoutMode = "HORIZONTAL";

  // Sidebar
  var sidebar = createSidebar();
  sidebar.layoutAlign = "STRETCH";

  // Main content area
  var mainContent = createFrame("Main Content", 1184, 900);
  mainContent.layoutMode = "VERTICAL";
  mainContent.fills = [];

  // Header
  var header = createHeader("Задачи");
  header.layoutAlign = "STRETCH";

  // Content
  var content = createFrame("Content", 1184, 844);
  setFill(content, COLORS.gray100);
  content.layoutMode = "VERTICAL";
  content.paddingLeft = 16;
  content.paddingRight = 16;
  content.paddingTop = 16;
  content.itemSpacing = 16;

  // Page title row
  var titleRow = createFrame("Title Row", 1152, 48);
  titleRow.layoutMode = "HORIZONTAL";
  titleRow.primaryAxisAlignItems = "SPACE_BETWEEN";
  titleRow.counterAxisAlignItems = "CENTER";
  titleRow.fills = [];

  var titleSection = createFrame("Title Section", 400, 48);
  titleSection.layoutMode = "VERTICAL";
  titleSection.itemSpacing = 4;
  titleSection.fills = [];

  var pageTitle = createText("Задачи", 20, "Bold", COLORS.gray900);

  var statsRow = createFrame("Stats", 300, 20);
  statsRow.layoutMode = "HORIZONTAL";
  statsRow.itemSpacing = 12;
  statsRow.fills = [];
  var stat1 = createText("24 задачи", 13, "Regular", COLORS.gray500);
  var stat2 = createText("89 SP", 13, "Medium", COLORS.blue600);
  statsRow.appendChild(stat1);
  statsRow.appendChild(stat2);

  titleSection.appendChild(pageTitle);
  titleSection.appendChild(statsRow);

  var createBtn = createButton("Создать", "primary", "sm");

  titleRow.appendChild(titleSection);
  titleRow.appendChild(createBtn);

  // Filter bar
  var filterBar = createFilterBar();

  // Kanban board
  var kanbanBoard = createFrame("Kanban Board", 1152, 600);
  kanbanBoard.layoutMode = "HORIZONTAL";
  kanbanBoard.itemSpacing = 16;
  kanbanBoard.fills = [];

  // Columns
  var backlogCol = createKanbanColumn("Backlog", 4, 12, COLORS.gray100, COLORS.gray400);
  var analysisCol = createKanbanColumn("Анализ", 2, 8, COLORS.purple50, COLORS.purple400);
  var devCol = createKanbanColumn("Разработка", 3, 21, COLORS.blue50, COLORS.blue500);
  var reviewCol = createKanbanColumn("Review", 1, 5, COLORS.yellow50, COLORS.yellow400);
  var testCol = createKanbanColumn("Тестирование", 2, 8, COLORS.orange50, COLORS.orange400);
  var doneCol = createKanbanColumn("Готово", 12, 35, COLORS.green50, COLORS.green500);

  kanbanBoard.appendChild(backlogCol);
  kanbanBoard.appendChild(analysisCol);
  kanbanBoard.appendChild(devCol);
  kanbanBoard.appendChild(reviewCol);
  kanbanBoard.appendChild(testCol);
  kanbanBoard.appendChild(doneCol);

  content.appendChild(titleRow);
  content.appendChild(filterBar);
  content.appendChild(kanbanBoard);

  mainContent.appendChild(header);
  mainContent.appendChild(content);

  page.appendChild(sidebar);
  page.appendChild(mainContent);

  return page;
}

function generateComponentLibrary() {
  var library = createFrame("EKF UI Components", 1200, 2000);
  setFill(library, COLORS.white);
  library.layoutMode = "VERTICAL";
  library.paddingLeft = 40;
  library.paddingTop = 40;
  library.itemSpacing = 40;

  // Title
  var title = createText("EKF Hub UI Kit", 32, "Bold", COLORS.gray900);
  library.appendChild(title);

  // Colors section
  var colorsSection = createFrame("Colors", 1120, 200);
  colorsSection.layoutMode = "VERTICAL";
  colorsSection.itemSpacing = 16;
  colorsSection.fills = [];

  var colorsTitle = createText("Colors", 20, "SemiBold", COLORS.gray800);
  colorsSection.appendChild(colorsTitle);

  var colorSwatches = createFrame("Color Swatches", 1120, 60);
  colorSwatches.layoutMode = "HORIZONTAL";
  colorSwatches.itemSpacing = 16;
  colorSwatches.fills = [];

  var colorList = [
    { name: "EKF Red", color: COLORS.ekfRed },
    { name: "EKF Dark", color: COLORS.ekfDark },
    { name: "Gray 100", color: COLORS.gray100 },
    { name: "Gray 200", color: COLORS.gray200 },
    { name: "Gray 500", color: COLORS.gray500 },
    { name: "Blue 500", color: COLORS.blue500 },
    { name: "Green 500", color: COLORS.green500 },
    { name: "Yellow 400", color: COLORS.yellow400 },
    { name: "Orange 400", color: COLORS.orange400 },
    { name: "Red 500", color: COLORS.red500 },
  ];

  colorList.forEach(function(c) {
    var swatch = createFrame(c.name, 100, 60);
    swatch.layoutMode = "VERTICAL";
    swatch.itemSpacing = 4;
    swatch.fills = [];

    var colorBox = createRoundedRect("Color", 100, 40, 8, c.color);
    var label = createText(c.name, 10, "Regular", COLORS.gray600);

    swatch.appendChild(colorBox);
    swatch.appendChild(label);
    colorSwatches.appendChild(swatch);
  });

  colorsSection.appendChild(colorSwatches);
  library.appendChild(colorsSection);

  // Buttons section
  var buttonsSection = createFrame("Buttons", 1120, 100);
  buttonsSection.layoutMode = "VERTICAL";
  buttonsSection.itemSpacing = 16;
  buttonsSection.fills = [];

  var buttonsTitle = createText("Buttons", 20, "SemiBold", COLORS.gray800);
  buttonsSection.appendChild(buttonsTitle);

  var buttonRow = createFrame("Button Row", 600, 50);
  buttonRow.layoutMode = "HORIZONTAL";
  buttonRow.itemSpacing = 16;
  buttonRow.fills = [];

  buttonRow.appendChild(createButton("Primary", "primary"));
  buttonRow.appendChild(createButton("Secondary", "secondary"));
  buttonRow.appendChild(createButton("Small", "primary", "sm"));

  buttonsSection.appendChild(buttonRow);
  library.appendChild(buttonsSection);

  // Inputs section
  var inputsSection = createFrame("Inputs", 1120, 100);
  inputsSection.layoutMode = "VERTICAL";
  inputsSection.itemSpacing = 16;
  inputsSection.fills = [];

  var inputsTitle = createText("Inputs & Selects", 20, "SemiBold", COLORS.gray800);
  inputsSection.appendChild(inputsTitle);

  var inputRow = createFrame("Input Row", 600, 50);
  inputRow.layoutMode = "HORIZONTAL";
  inputRow.itemSpacing = 16;
  inputRow.fills = [];

  inputRow.appendChild(createInput("Поиск задач...", 200));
  inputRow.appendChild(createSelect("Все проекты", 150));

  inputsSection.appendChild(inputRow);
  library.appendChild(inputsSection);

  // Task Cards section
  var cardsSection = createFrame("Task Cards", 1120, 200);
  cardsSection.layoutMode = "VERTICAL";
  cardsSection.itemSpacing = 16;
  cardsSection.fills = [];

  var cardsTitle = createText("Task Cards", 20, "SemiBold", COLORS.gray800);
  cardsSection.appendChild(cardsTitle);

  var cardsRow = createFrame("Cards Row", 1100, 140);
  cardsRow.layoutMode = "HORIZONTAL";
  cardsRow.itemSpacing = 16;
  cardsRow.fills = [];

  cardsRow.appendChild(createTaskCard("TASK-138", "Исправить отображение участников встречи", "critical", 5, "Д"));
  cardsRow.appendChild(createTaskCard("TASK-139", "Workflow транскрибирования", "medium", 8, "И"));
  cardsRow.appendChild(createTaskCard("TASK-140", "Service Desk ITIL спецификация", "high", 13, "А"));

  cardsSection.appendChild(cardsRow);
  library.appendChild(cardsSection);

  // Sidebar
  var sidebarSection = createFrame("Sidebar", 1120, 500);
  sidebarSection.layoutMode = "VERTICAL";
  sidebarSection.itemSpacing = 16;
  sidebarSection.fills = [];

  var sidebarTitle = createText("Sidebar", 20, "SemiBold", COLORS.gray800);
  sidebarSection.appendChild(sidebarTitle);

  var sidebar = createSidebar();
  sidebarSection.appendChild(sidebar);
  library.appendChild(sidebarSection);

  return library;
}

// =====================
// MAIN
// =====================
figma.showUI(__html__, { width: 320, height: 400 });

figma.ui.onmessage = function(msg) {
  loadFonts().then(function() {
    if (msg.type === 'generate-tasks-page') {
      var page = generateTasksPage();
      figma.currentPage.appendChild(page);
      figma.viewport.scrollAndZoomIntoView([page]);
      figma.notify('Tasks page generated!');
    }

    if (msg.type === 'generate-components') {
      var library = generateComponentLibrary();
      figma.currentPage.appendChild(library);
      figma.viewport.scrollAndZoomIntoView([library]);
      figma.notify('Component library generated!');
    }

    if (msg.type === 'generate-sidebar') {
      var sidebar = createSidebar();
      figma.currentPage.appendChild(sidebar);
      figma.viewport.scrollAndZoomIntoView([sidebar]);
      figma.notify('Sidebar generated!');
    }

    if (msg.type === 'generate-header') {
      var header = createHeader();
      figma.currentPage.appendChild(header);
      figma.viewport.scrollAndZoomIntoView([header]);
      figma.notify('Header generated!');
    }

    if (msg.type === 'generate-task-card') {
      var card = createTaskCard("TASK-001", "Example Task", "medium", 5, "А");
      figma.currentPage.appendChild(card);
      figma.viewport.scrollAndZoomIntoView([card]);
      figma.notify('Task card generated!');
    }

    if (msg.type === 'generate-kanban-column') {
      var column = createKanbanColumn("Backlog", 4, 12, COLORS.gray100, COLORS.gray400);
      figma.currentPage.appendChild(column);
      figma.viewport.scrollAndZoomIntoView([column]);
      figma.notify('Kanban column generated!');
    }

    if (msg.type === 'generate-filter-bar') {
      var filterBar = createFilterBar();
      figma.currentPage.appendChild(filterBar);
      figma.viewport.scrollAndZoomIntoView([filterBar]);
      figma.notify('Filter bar generated!');
    }
  });
};
