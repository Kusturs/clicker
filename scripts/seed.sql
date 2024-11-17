TRUNCATE TABLE banners CASCADE;

INSERT INTO banners (id, name) VALUES 
(1, 'Banner #1 - Homepage Hero'),
(2, 'Banner #2 - Sidebar Promo'),
(3, 'Banner #3 - Footer Ad'),
(4, 'Banner #4 - Product Page Top'),
(5, 'Banner #5 - Category Showcase'),
(6, 'Banner #6 - Mobile App Promo'),
(7, 'Banner #7 - Newsletter Signup'),
(8, 'Banner #8 - Special Offer'),
(9, 'Banner #9 - Holiday Campaign'),
(10, 'Banner #10 - Flash Sale'),
(11, 'Banner #11 - Blog Sidebar'),
(12, 'Banner #12 - Search Results'),
(13, 'Banner #13 - Account Page'),
(14, 'Banner #14 - Checkout Upsell'),
(15, 'Banner #15 - Social Media'),
(16, 'Banner #16 - Email Campaign'),
(17, 'Banner #17 - Partner Promo'),
(18, 'Banner #18 - Seasonal Deal'),
(19, 'Banner #19 - Limited Time'),
(20, 'Banner #20 - Member Exclusive');

WITH series AS (
  SELECT generate_series(21, 100) as id
)
INSERT INTO banners (id, name)
SELECT 
    id,
    'Banner #' || id::text || ' - ' || 
    CASE (id % 5)
        WHEN 0 THEN 'Premium Ad'
        WHEN 1 THEN 'Featured Promotion'
        WHEN 2 THEN 'Special Campaign'
        WHEN 3 THEN 'Seasonal Offer'
        WHEN 4 THEN 'Custom Deal'
    END as name
FROM series;
