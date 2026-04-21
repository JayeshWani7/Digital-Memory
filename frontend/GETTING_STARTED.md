# Digital Memory Frontend - Getting Started

This is a production-ready Next.js frontend for the Digital Memory semantic search engine.

## Quick Setup

### 1. Install Dependencies
```bash
cd frontend
npm install
```

### 2. Start Development Server
```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### 3. Build for Production
```bash
npm run build
npm run start
```

## Key Features

✅ **Semantic Search** - Natural language queries over your knowledge base
✅ **Responsive Design** - Mobile-first UI that works everywhere
✅ **Real-time Caching** - Fast, efficient data fetching with React Query
✅ **TypeScript** - Type-safe development
✅ **Tailwind CSS** - Modern, utility-first styling

## Project Structure

- **`app/`** - Next.js pages and routes
  - `page.tsx` - Home/search page
  - `knowledge/page.tsx` - Browse all knowledge
  - `results/[id]/page.tsx` - Result detail view
- **`components/`** - Reusable React components
  - `SearchBar.tsx` - Search input
  - `ResultCard.tsx` - Individual result card
  - `ResultsList.tsx` - Results grid
- **`lib/`** - Utilities and types
  - `api.ts` - API client
  - `types.ts` - TypeScript types

## Environment Variables

Create `.env.local` (already created):
```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8000
```

For production, update `.env.production` with your backend API URL.

## API Endpoints Used

- `POST /api/v1/query` - Semantic search
- `GET /api/v1/knowledge/:id` - Get single result
- `GET /api/v1/knowledge` - List all knowledge
- `GET /health` - Health check
- `GET /status` - Service status

See `../docs/API.md` for full API documentation.

## Tech Stack

- **Next.js 14** - React framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **React Query** - Data fetching
- **Axios** - HTTP client
- **Zustand** - State management (optional)

## Customization Ideas

1. **Add Filters** - Source, author, date range UI
2. **Dark Mode** - Toggle between light/dark themes
3. **Bookmarks** - Save favorite results
4. **Export** - Download results as PDF/JSON
5. **Analytics** - Track search patterns
6. **Embeddings** - Visualize result similarity

## Troubleshooting

### API Connection Issues
- Ensure backend is running on `http://localhost:8000`
- Check `NEXT_PUBLIC_API_BASE_URL` in `.env.local`
- Verify CORS headers in backend

### Build Issues
- Run `npm install` again
- Clear `.next` folder: `rm -rf .next`
- Check Node.js version: `node --version` (need 18+)

## Next Steps

1. Run `npm install` to install dependencies
2. Start the backend services (see `../QUICKSTART.md`)
3. Run `npm run dev` to start the frontend
4. Open http://localhost:3000 and start searching!

Happy coding! 🚀
