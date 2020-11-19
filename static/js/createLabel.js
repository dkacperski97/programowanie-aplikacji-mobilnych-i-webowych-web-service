import attachEvents from './validationHelper.js';
import labelValidations from './labelValidations.js';

attachEvents(labelValidations, 'createLabelForm', true);
