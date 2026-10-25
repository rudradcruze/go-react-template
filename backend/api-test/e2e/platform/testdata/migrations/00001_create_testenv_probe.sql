-- +goose Up
CREATE TABLE testenv_probe (id int);

-- +goose Down
DROP TABLE testenv_probe;
