# Deployment Guide

## Prerequisites

1. **GitHub Account** - For code repository
2. **Vercel Account** - For hosting (free tier available)
3. **Supabase Project** - For PostgreSQL database

---

## Step 1: Setup Supabase Database

1. Go to [supabase.com](https://supabase.com) and create a project
2. Wait for the project to be ready
3. Go to SQL Editor and run the schema (already set up via v0)
4. Copy your `DATABASE_URL` from Project Settings → Database → Connection strings

---

## Step 2: Create GitHub Repository

1. Create a new private repository on GitHub
2. Clone locally:
   \`\`\`bash
   git clone <your-repo-url>
   cd healthy-lifestyle-api
   \`\`\`
3. Add all Go files:
   \`\`\`bash
   git add .
   git commit -m "Initial Go backend setup"
   git push origin main
   \`\`\`

---

## Step 3: Deploy to Vercel

### Option A: Using Vercel Dashboard (Recommended)

1. Go to [vercel.com](https://vercel.com) and sign up
2. Click "New Project"
3. Select your GitHub repository
4. Vercel will auto-detect Go
5. Click "Environment Variables" and add:
   - `DATABASE_URL` = Your Supabase connection string
   - `JWT_SECRET` = Generate a random string (use: `openssl rand -base64 32`)
6. Click "Deploy"

### Option B: Using Vercel CLI

\`\`\`bash
# Install Vercel CLI
npm i -g vercel

# Login to Vercel
vercel login

# Deploy
vercel

# Add environment variables
vercel env add DATABASE_URL
vercel env add JWT_SECRET

# Redeploy with env vars
vercel --prod
\`\`\`

---

## Step 4: Test Your API

Once deployed, you'll get a URL like: `https://your-project.vercel.app`

### Test Register:
\`\`\`bash
curl -X POST https://your-project.vercel.app/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
\`\`\`

### Test Login:
\`\`\`bash
curl -X POST https://your-project.vercel.app/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
\`\`\`

### Test Profile (use token from login):
\`\`\`bash
curl -X GET https://your-project.vercel.app/api/profile \
  -H "Authorization: Bearer <your-token>"
\`\`\`

---

## Troubleshooting

### Issue: "DATABASE_URL not set"
- Check Environment Variables in Vercel Dashboard
- Make sure you've added `DATABASE_URL` before deploying
- Redeploy after adding variables

### Issue: "Failed to connect to database"
- Verify `DATABASE_URL` is correct
- Check if Supabase project is running
- Test connection locally with `.env.example`

### Issue: "Invalid token"
- Make sure `JWT_SECRET` matches between register and login
- Check token expiration (24 hours)
- Verify Authorization header format: `Bearer <token>`

---

## Local Development

1. Create `.env` file (based on `.env.example`):
   \`\`\`
   DATABASE_URL=postgresql://user:password@localhost:5432/db
   JWT_SECRET=your-secret-key
   PORT=8080
   \`\`\`

2. Install dependencies:
   \`\`\`bash
   go mod download
   \`\`\`

3. Run locally:
   \`\`\`bash
   go run ./cmd/main.go
   \`\`\`

4. Test endpoints at `http://localhost:8080`

---

## Next Steps

- Add more endpoints as needed
- Implement rate limiting
- Add request logging
- Setup CI/CD pipeline
- Add unit tests
