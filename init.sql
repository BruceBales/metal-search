-- Create the bands table
CREATE TABLE bands (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    formed_in INT,
    status VARCHAR(255),
    years_active VARCHAR(255),
    genre VARCHAR(255),
    themes VARCHAR(255),
    label VARCHAR(255),
    band_cover VARCHAR(255),
    spotify_link VARCHAR(255)
);

-- Create the albums table
CREATE TABLE albums (
    id INT AUTO_INCREMENT PRIMARY KEY,
    band_id INT,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255),
    year INT,
    link VARCHAR(255),
    FOREIGN KEY (band_id) REFERENCES bands(id) ON DELETE CASCADE
);

CREATE TABLE `hits` (
  `id` int NOT NULL AUTO_INCREMENT,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `ip_address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `user_agent` varchar(255) DEFAULT NULL,
  `path` varchar(255) DEFAULT NULL,
  `params` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
);

-- Insert sample data into bands table
INSERT INTO bands (name, country, location, formed_in, status, years_active, genre, themes, label, band_cover, spotify_link)
VALUES
('Amorphis', 'Finland', 'Helsinki', 1990, 'Active', '1990-present', 'Progressive Metal', 'Mythology, History', 'Nuclear Blast', 'https://example.com/amorphis.jpg', 'https://open.spotify.com/artist/4S2qftLTvdEFvIPPTYmeg6');

-- Insert sample data into albums table
INSERT INTO albums (band_id, name, type, year, link)
VALUES
(1, 'Tales from the Thousand Lakes', 'Album', 1994, 'https://example.com/tales.jpg'),
(1, 'Queen of Time', 'Album', 2018, 'https://example.com/queen.jpg');

