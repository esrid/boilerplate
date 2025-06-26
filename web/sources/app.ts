// src/application.js
import { Application } from "@hotwired/stimulus";

// Import new modular controllers
import TreeVisualizationController from "./controllers/tree_visualization_controller.ts";
import NodeDetailsController from "./controllers/node_details_controller.ts";
import NodeEditController from "./controllers/node_edit_controller.ts";
import TreeCoordinatorController from "./controllers/tree_coordinator_controller.ts";
import SidebarManagerController from "./controllers/sidebar_manager_controller.ts";
import ThemeController from "./controllers/theme_controller.ts";

import loading_controller from "./controllers/loading_controller.ts";

// Import legacy controller for backward compatibility
import CanvaController from "./controllers/canva_controller.ts";

import GoalFormController from "./controllers/goal_form_controller.ts";

//@ts-ignore
window.Stimulus = Application.start();

// Register new modular controllers
//@ts-ignore
Stimulus.register("tree-visualization", TreeVisualizationController);
//@ts-ignore
Stimulus.register("node-details", NodeDetailsController);
//@ts-ignore
Stimulus.register("node-edit", NodeEditController);
//@ts-ignore
Stimulus.register("tree-coordinator", TreeCoordinatorController);
//@ts-ignore
Stimulus.register("sidebar-manager", SidebarManagerController);
//@ts-ignore
Stimulus.register("theme", ThemeController);

//@ts-ignore
Stimulus.register("loading", loading_controller);

// Register legacy controller for backward compatibility
//@ts-ignore
Stimulus.register("canvas", CanvaController);

//@ts-ignore
Stimulus.register("goal-form", GoalFormController);
