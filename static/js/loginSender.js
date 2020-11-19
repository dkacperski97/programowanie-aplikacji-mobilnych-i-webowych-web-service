import attachEvents from "./validationHelper.js";
import userValidations from "./userValidations.js";

attachEvents(userValidations, 'loginForm', true);
