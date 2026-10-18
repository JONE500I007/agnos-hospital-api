CREATE TABLE hospitals (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE staff (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    hospital_id   INT  NOT NULL REFERENCES hospitals(id)
);

CREATE TABLE patients (
    id             SERIAL PRIMARY KEY,
    hospital_id    INT  NOT NULL REFERENCES hospitals(id),
    first_name_th  TEXT,
    middle_name_th TEXT,
    last_name_th   TEXT,
    first_name_en  TEXT,
    middle_name_en TEXT,
    last_name_en   TEXT,
    date_of_birth  DATE,
    patient_hn     TEXT,
    national_id    TEXT,
    passport_id    TEXT,
    phone_number   TEXT,
    email          TEXT,
    gender         CHAR(1) CHECK (gender IN ('M', 'F'))
);

CREATE INDEX idx_patients_national_id ON patients(national_id);
CREATE INDEX idx_patients_passport_id ON patients(passport_id);
CREATE INDEX idx_patients_hospital_id ON patients(hospital_id);

INSERT INTO hospitals (name) VALUES ('Hospital A'), ('Hospital B');

INSERT INTO patients (hospital_id, first_name_th, last_name_th, first_name_en, last_name_en,
                      date_of_birth, patient_hn, national_id, phone_number, email, gender)
VALUES
 (1, 'สมชาย', 'ใจดี', 'Somchai', 'Jaidee', '1990-05-12', 'HN0001', '1100100000011', '0810000001', 'somchai@example.com', 'M'),
 (1, 'สมหญิง', 'รักดี', 'Somying', 'Rakdee', '1985-09-30', 'HN0002', '1100100000022', '0810000002', 'somying@example.com', 'F'),
 (2, 'วิชัย', 'มั่นคง', 'Wichai', 'Mankong', '1978-01-20', 'HN0003', '1100100000033', '0810000003', 'wichai@example.com', 'M');