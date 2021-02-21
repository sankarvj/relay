What is this choice expression?

The expressions in the choices will be evaluated and if the evalution returns true then that corresponding choiceID will be
set for that field.
example: status field, the status will set to overdue if the due_by field value is lesser than the current time. Here the 
expression would be like: if due_by_time < current_time --> due_by_choice_id