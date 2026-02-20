-- Create a table for Departments
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    location VARCHAR(100)
);

-- Create a table for Employees
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE,
    hire_date DATE DEFAULT CURRENT_DATE,
    salary NUMERIC(12, 2),
    dept_id INTEGER REFERENCES departments(id) ON DELETE SET NULL,
    metadata JSONB -- Testing Postgres JSON capabilities
);

-- Create a view for quick reporting
CREATE VIEW employee_directory AS
SELECT 
    e.id, 
    e.first_name || ' ' || e.last_name AS full_name, 
    d.name AS department, 
    e.email
FROM employees e
LEFT JOIN departments d ON e.dept_id = d.id;

INSERT INTO departments (name, location) VALUES
('Engineering', 'New York'),
('Sales', 'Chicago'),
('Marketing', 'San Francisco'),
('Legal', 'Washington D.C.');

-- Insert Employees with some JSON metadata
INSERT INTO employees (first_name, last_name, email, salary, dept_id, metadata) VALUES
('Alice', 'Smith', 'alice@example.com', 125000, 1, '{"skills": ["Go", "Kubernetes", "Postgres"], "remote": true}'),
('Bob', 'Johnson', 'bob@example.com', 98000, 1, '{"skills": ["Python", "Docker"], "remote": false}'),
('Charlie', 'Brown', 'charlie@example.com', 85000, 2, '{"skills": ["Negotiation", "CRM"], "years_exp": 5}'),
('Diana', 'Prince', 'diana@example.com', 110000, 3, '{"skills": ["SEO", "Copywriting"], "socials": ["Twitter", "LinkedIn"]}'),
('Edward', 'Norton', 'edward@example.com', 150000, 4, '{"skills": ["Litigation", "Compliance"]}');

-- Add a column for tracking
ALTER TABLE employees ADD COLUMN last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Create the function
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create the trigger
CREATE TRIGGER update_employee_modtime
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE PROCEDURE update_modified_column();