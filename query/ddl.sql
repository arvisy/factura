CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    deposit INT NOT NULL,
    role VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS equipments (
    equipment_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    stock INT NOT NULL,
    rental_cost INT NOT NULL,
    category VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS rents (
    rent_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    payment_date VARCHAR(50),
    payment_status VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS topups (
    topup_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    topup_amount INT NOT NULL,
    topup_date VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS rent_equipments (
    rent_equipment_id SERIAL PRIMARY KEY,
    rent_id INT REFERENCES rents(rent_id),
    equipment_id INT REFERENCES equipments(equipment_id),
    quantity INT NOT NULL,
    start_date VARCHAR(50) NOT NULL,
    end_date VARCHAR(50) NOT NULL,
    total_rental_cost INT NOT NULL
);


CREATE TABLE IF NOT EXISTS payment (
    payment_id SERIAL PRIMARY KEY,
    payment_method VARCHAR(50) NOT NULL
);