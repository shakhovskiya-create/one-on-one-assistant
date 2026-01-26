// EKF Hub UI Kit - Figma Plugin
// Generates UI components in EKF Hub design style

// =====================
// COLOR PALETTE
// =====================
var COLORS = {
  ekfRed: { r: 229/255, g: 57/255, b: 53/255 },
  ekfDark: { r: 26/255, g: 26/255, b: 46/255 },
  white: { r: 1, g: 1, b: 1 },
  gray100: { r: 243/255, g: 244/255, b: 246/255 },
  gray200: { r: 229/255, g: 231/255, b: 235/255 },
  gray300: { r: 209/255, g: 213/255, b: 219/255 },
  gray400: { r: 156/255, g: 163/255, b: 175/255 },
  gray500: { r: 107/255, g: 114/255, b: 128/255 },
  gray600: { r: 75/255, g: 85/255, b: 99/255 },
  gray700: { r: 55/255, g: 65/255, b: 81/255 },
  gray800: { r: 31/255, g: 41/255, b: 55/255 },
  gray900: { r: 17/255, g: 24/255, b: 39/255 },
  blue50: { r: 239/255, g: 246/255, b: 255/255 },
  blue500: { r: 59/255, g: 130/255, b: 246/255 },
  blue600: { r: 37/255, g: 99/255, b: 235/255 },
  green50: { r: 240/255, g: 253/255, b: 244/255 },
  green500: { r: 34/255, g: 197/255, b: 94/255 },
  yellow50: { r: 254/255, g: 252/255, b: 232/255 },
  yellow400: { r: 250/255, g: 204/255, b: 21/255 },
  orange50: { r: 255/255, g: 247/255, b: 237/255 },
  orange400: { r: 251/255, g: 146/255, b: 60/255 },
  orange600: { r: 234/255, g: 88/255, b: 12/255 },
  purple50: { r: 250/255, g: 245/255, b: 255/255 },
  purple400: { r: 192/255, g: 132/255, b: 252/255 },
  red500: { r: 239/255, g: 68/255, b: 68/255 },
};

// Font - using Roboto which is always available in Figma
var FONT_REGULAR = { family: "Roboto", style: "Regular" };
var FONT_MEDIUM = { family: "Roboto", style: "Medium" };
var FONT_BOLD = { family: "Roboto", style: "Bold" };

function loadFonts() {
  return Promise.all([
    figma.loadFontAsync(FONT_REGULAR),
    figma.loadFontAsync(FONT_MEDIUM),
    figma.loadFontAsync(FONT_BOLD)
  ]);
}

function createFrame(name, width, height) {
  var frame = figma.createFrame();
  frame.name = name;
  frame.resize(width, height);
  frame.fills = [];
  return frame;
}

function setFill(node, color) {
  node.fills = [{ type: 'SOLID', color: color }];
}

function createText(content, fontSize, fontName, color) {
  if (!fontName) fontName = FONT_REGULAR;
  if (!color) color = COLORS.gray900;
  var text = figma.createText();
  text.fontName = fontName;
  text.characters = content;
  text.fontSize = fontSize;
  text.fills = [{ type: 'SOLID', color: color }];
  return text;
}

function createRect(name, width, height, radius, color) {
  var rect = figma.createRectangle();
  rect.name = name;
  rect.resize(width, height);
  rect.cornerRadius = radius;
  setFill(rect, color);
  return rect;
}

// =====================
// SIMPLE COMPONENTS
// =====================

function createSidebar() {
  var sidebar = createFrame("Sidebar", 240, 600);
  setFill(sidebar, COLORS.ekfDark);

  // Logo
  var logo = createRect("Logo", 40, 28, 4, COLORS.ekfRed);
  logo.x = 16;
  logo.y = 16;
  sidebar.appendChild(logo);

  var logoText = createText("EKF", 14, FONT_BOLD, COLORS.white);
  logoText.x = 22;
  logoText.y = 20;
  sidebar.appendChild(logoText);

  var hubText = createText("Hub", 16, FONT_MEDIUM, COLORS.white);
  hubText.x = 64;
  hubText.y = 18;
  sidebar.appendChild(hubText);

  // Menu items
  var menuItems = ["Dashboard", "Employees", "Tasks", "Meetings", "Service Desk", "Messages"];
  var yPos = 80;

  for (var i = 0; i < menuItems.length; i++) {
    var isActive = (i === 2); // Tasks is active

    if (isActive) {
      var activeBg = createRect("Active BG", 208, 40, 8, COLORS.ekfRed);
      activeBg.x = 16;
      activeBg.y = yPos;
      sidebar.appendChild(activeBg);
    }

    var menuText = createText(menuItems[i], 14, FONT_REGULAR, isActive ? COLORS.white : COLORS.gray300);
    menuText.x = 48;
    menuText.y = yPos + 12;
    sidebar.appendChild(menuText);

    yPos += 48;
  }

  return sidebar;
}

function createHeader() {
  var header = createFrame("Header", 1000, 56);
  setFill(header, COLORS.white);

  // Border bottom
  var border = createRect("Border", 1000, 1, 0, COLORS.gray200);
  border.x = 0;
  border.y = 55;
  header.appendChild(border);

  // Title
  var title = createText("Tasks", 20, FONT_BOLD, COLORS.gray800);
  title.x = 24;
  title.y = 16;
  header.appendChild(title);

  // User avatar
  var avatar = createRect("Avatar", 32, 32, 16, COLORS.ekfRed);
  avatar.x = 944;
  avatar.y = 12;
  header.appendChild(avatar);

  var avatarText = createText("A", 14, FONT_MEDIUM, COLORS.white);
  avatarText.x = 955;
  avatarText.y = 18;
  header.appendChild(avatarText);

  return header;
}

function createTaskCard() {
  var card = createFrame("Task Card", 260, 100);
  setFill(card, COLORS.white);
  card.cornerRadius = 8;
  card.strokes = [{ type: 'SOLID', color: COLORS.gray200 }];
  card.strokeWeight = 1;
  card.strokeLeftWeight = 3;
  card.strokes = [{ type: 'SOLID', color: COLORS.yellow400 }];

  // Task ID
  var taskId = createText("TASK-001", 11, FONT_REGULAR, COLORS.gray400);
  taskId.x = 12;
  taskId.y = 12;
  card.appendChild(taskId);

  // SP Badge
  var spBadge = createRect("SP Badge", 32, 18, 4, COLORS.gray100);
  spBadge.x = 216;
  spBadge.y = 10;
  card.appendChild(spBadge);

  var spText = createText("5 SP", 10, FONT_MEDIUM, COLORS.gray600);
  spText.x = 221;
  spText.y = 12;
  card.appendChild(spText);

  // Title
  var titleText = createText("Example task title here", 13, FONT_MEDIUM, COLORS.gray900);
  titleText.x = 12;
  titleText.y = 40;
  card.appendChild(titleText);

  // Assignee
  var assignee = createRect("Assignee", 24, 24, 12, COLORS.ekfRed);
  assignee.x = 224;
  assignee.y = 68;
  card.appendChild(assignee);

  return card;
}

function createKanbanColumn() {
  var column = createFrame("Kanban Column", 280, 500);
  setFill(column, COLORS.gray100);
  column.cornerRadius = 8;

  // Header
  var dot = createRect("Dot", 8, 8, 4, COLORS.blue500);
  dot.x = 12;
  dot.y = 20;
  column.appendChild(dot);

  var colName = createText("In Progress", 13, FONT_MEDIUM, COLORS.gray800);
  colName.x = 28;
  colName.y = 14;
  column.appendChild(colName);

  var countBadge = createRect("Count", 24, 20, 10, COLORS.gray200);
  countBadge.x = 110;
  countBadge.y = 14;
  column.appendChild(countBadge);

  var countText = createText("3", 11, FONT_MEDIUM, COLORS.gray600);
  countText.x = 118;
  countText.y = 17;
  column.appendChild(countText);

  var spText = createText("21 SP", 11, FONT_REGULAR, COLORS.gray500);
  spText.x = 236;
  spText.y = 17;
  column.appendChild(spText);

  return column;
}

function createFilterBar() {
  var bar = createFrame("Filter Bar", 900, 48);
  setFill(bar, COLORS.white);
  bar.cornerRadius = 8;
  bar.strokes = [{ type: 'SOLID', color: COLORS.gray200 }];
  bar.strokeWeight = 1;

  // Search
  var searchBg = createRect("Search", 180, 32, 6, COLORS.gray100);
  searchBg.x = 8;
  searchBg.y = 8;
  bar.appendChild(searchBg);

  var searchText = createText("Search tasks...", 13, FONT_REGULAR, COLORS.gray400);
  searchText.x = 20;
  searchText.y = 15;
  bar.appendChild(searchText);

  // Filter buttons
  var filters = ["All Projects", "All Sprints", "All Assignees"];
  var xPos = 200;

  for (var i = 0; i < filters.length; i++) {
    var filterBg = createRect("Filter " + i, 100, 32, 6, COLORS.white);
    filterBg.x = xPos;
    filterBg.y = 8;
    filterBg.strokes = [{ type: 'SOLID', color: COLORS.gray200 }];
    filterBg.strokeWeight = 1;
    bar.appendChild(filterBg);

    var filterText = createText(filters[i], 12, FONT_REGULAR, COLORS.gray600);
    filterText.x = xPos + 10;
    filterText.y = 16;
    bar.appendChild(filterText);

    xPos += 112;
  }

  // View toggle
  var kanbanBtn = createRect("Kanban Btn", 70, 32, 6, COLORS.ekfRed);
  kanbanBtn.x = 822;
  kanbanBtn.y = 8;
  bar.appendChild(kanbanBtn);

  var kanbanText = createText("Kanban", 12, FONT_MEDIUM, COLORS.white);
  kanbanText.x = 834;
  kanbanText.y = 16;
  bar.appendChild(kanbanText);

  return bar;
}

function generateTasksPage() {
  var page = createFrame("Tasks Page", 1440, 900);
  setFill(page, COLORS.gray100);

  // Sidebar
  var sidebar = createSidebar();
  sidebar.x = 0;
  sidebar.y = 0;
  sidebar.resize(240, 900);
  page.appendChild(sidebar);

  // Header
  var header = createHeader();
  header.x = 240;
  header.y = 0;
  header.resize(1200, 56);
  page.appendChild(header);

  // Title
  var titleText = createText("Tasks", 24, FONT_BOLD, COLORS.gray900);
  titleText.x = 264;
  titleText.y = 80;
  page.appendChild(titleText);

  // Create button
  var createBtn = createRect("Create Btn", 80, 32, 6, COLORS.ekfRed);
  createBtn.x = 1336;
  createBtn.y = 76;
  page.appendChild(createBtn);

  var createText = createText("Create", 13, FONT_MEDIUM, COLORS.white);
  createText.x = 1354;
  createText.y = 84;
  page.appendChild(createText);

  // Filter bar
  var filterBar = createFilterBar();
  filterBar.x = 264;
  filterBar.y = 130;
  filterBar.resize(1152, 48);
  page.appendChild(filterBar);

  // Kanban columns
  var columns = [
    { name: "Backlog", color: COLORS.gray400, count: 4 },
    { name: "Analysis", color: COLORS.purple400, count: 2 },
    { name: "Development", color: COLORS.blue500, count: 3 },
    { name: "Review", color: COLORS.yellow400, count: 1 },
    { name: "Done", color: COLORS.green500, count: 8 }
  ];

  var colX = 264;
  for (var i = 0; i < columns.length; i++) {
    var col = createFrame("Column " + columns[i].name, 220, 680);
    setFill(col, COLORS.gray100);
    col.cornerRadius = 8;
    col.x = colX;
    col.y = 200;

    // Dot
    var dot = createRect("Dot", 8, 8, 4, columns[i].color);
    dot.x = 12;
    dot.y = 16;
    col.appendChild(dot);

    // Name
    var colName = createText(columns[i].name, 13, FONT_MEDIUM, COLORS.gray800);
    colName.x = 28;
    colName.y = 10;
    col.appendChild(colName);

    // Count
    var countText = createText(columns[i].count.toString(), 11, FONT_MEDIUM, COLORS.gray500);
    countText.x = 190;
    countText.y = 12;
    col.appendChild(countText);

    page.appendChild(col);
    colX += 228;
  }

  return page;
}

// =====================
// MAIN
// =====================
figma.showUI(__html__, { width: 320, height: 480 });

figma.ui.onmessage = function(msg) {
  loadFonts().then(function() {
    try {
      var node;

      if (msg.type === 'generate-tasks-page') {
        node = generateTasksPage();
      } else if (msg.type === 'generate-sidebar') {
        node = createSidebar();
      } else if (msg.type === 'generate-header') {
        node = createHeader();
      } else if (msg.type === 'generate-filter-bar') {
        node = createFilterBar();
      } else if (msg.type === 'generate-kanban-column') {
        node = createKanbanColumn();
      } else if (msg.type === 'generate-task-card') {
        node = createTaskCard();
      } else if (msg.type === 'generate-components') {
        // Generate all components
        var y = 0;

        var sidebar = createSidebar();
        sidebar.y = y;
        figma.currentPage.appendChild(sidebar);
        y += 620;

        var header = createHeader();
        header.y = y;
        figma.currentPage.appendChild(header);
        y += 80;

        var filterBar = createFilterBar();
        filterBar.y = y;
        figma.currentPage.appendChild(filterBar);
        y += 80;

        var column = createKanbanColumn();
        column.y = y;
        figma.currentPage.appendChild(column);

        var card = createTaskCard();
        card.x = 300;
        card.y = y;
        figma.currentPage.appendChild(card);

        figma.notify('All components generated!');
        figma.viewport.scrollAndZoomIntoView([sidebar, header, filterBar, column, card]);
        return;
      }

      if (node) {
        figma.currentPage.appendChild(node);
        figma.viewport.scrollAndZoomIntoView([node]);
        figma.notify('Generated: ' + node.name);
      }
    } catch (e) {
      figma.notify('Error: ' + e.message, { error: true });
    }
  }).catch(function(e) {
    figma.notify('Font error: ' + e.message, { error: true });
  });
};
