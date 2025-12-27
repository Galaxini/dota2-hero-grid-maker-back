# dota2-hero-grid-maker-back

### Manual verification

- Register user: POST /auth/register
- Login: POST /auth/login
- Create grid (auth): POST /grids
- Fetch user grids (auth): GET /grids
- Fetch default grid (public): GET /default-grid

The default grid is stored in the database as JSONB and loaded via GET /default-grid.
