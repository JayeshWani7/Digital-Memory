# Digital Memory Frontend

Production-ready Next.js frontend for the Digital Memory semantic search engine.

## 📋 Features

- **Semantic Search** - Natural language queries over your knowledge base
- **Source Filtering** - Filter results by Slack, GitHub, author, daterange
- **Responsive Design** - Mobile-first, works on all screen sizes
- **Real-time Results** - Instant search with React Query caching
- **Result Detail Pages** - Full context for each knowledge item
- **Knowledge Browser** - Browse and paginate through all indexed items

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ and npm/yarn/pnpm
- Digital Memory backend running on `http://localhost:8000`

### Installation

```bash
cd frontend
npm install
```

### Development

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build for Production

```bash
npm run build
npm run start
```

## 🗂️ Project Structure

```
frontend/
├── app/                    # Next.js App Router
│   ├── layout.tsx         # Root layout with navigation
│   ├── page.tsx           # Home/search page
│   ├── globals.css        # Global styles
│   ├── knowledge/
│   │   └── page.tsx       # Knowledge browser
│   └── results/
│       └── [id]/
│           └── page.tsx   # Result detail view
├── components/            # Reusable React components
│   ├── SearchBar.tsx      # Search input component
│   ├── ResultCard.tsx     # Individual result card
│   └── ResultsList.tsx    # Grid of results
├── lib/
│   ├── api.ts            # API client wrapper
│   └── types.ts          # TypeScript types
├── public/               # Static assets
├── package.json
├── tsconfig.json
├── tailwind.config.ts    # Tailwind CSS config
└── next.config.js        # Next.js config
```

## 🛠️ Tech Stack

- **Framework**: [Next.js 14](https://nextjs.org/) - React with server-side rendering
- **Language**: [TypeScript](https://www.typescriptlang.org/) - Type-safe development
- **Styling**: [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS
- **HTTP Client**: [Axios](https://axios-http.com/) - Promise-based HTTP client
- **Data Fetching**: [@tanstack/react-query](https://tanstack.com/query) - Server state management
- **State Management**: [Zustand](https://github.com/pmndrs/zustand) - Lightweight state management
- **UI Icons**: [@radix-ui/react-icons](https://www.radix-ui.com/icons) - Accessible icon set

## 📡 API Integration

The frontend communicates with the Digital Memory backend API:

```typescript
// Search results
POST /api/v1/query
{
  "query": "What database decisions were made?",
  "top_k": 10,
  "filters": {}
}

// Get single result
GET /api/v1/knowledge/:id

// List all knowledge
GET /api/v1/knowledge?page=1&limit=20

// Health check
GET /health

// Service status
GET /status
```

See [docs/API.md](../docs/API.md) for complete API documentation.

## 🎨 Customization

### Update API Base URL

Edit `.env.local`:
```env
NEXT_PUBLIC_API_BASE_URL=http://your-api-host:8000
```

### Styling

Tailwind CSS is configured in `tailwind.config.ts`. Customize colors, fonts, and spacing there.

### Components

All React components are in the `components/` directory and are ready to customize.

## 🧪 Testing (Optional)

Add testing dependencies:
```bash
npm install --save-dev @testing-library/react @testing-library/jest-dom vitest
```

Then create `.test.tsx` files next to components.

## 📦 Deployment

### Vercel (Recommended for Next.js)

1. Push code to GitHub
2. Connect repo to [Vercel](https://vercel.com)
3. Set `NEXT_PUBLIC_API_BASE_URL` environment variable
4. Deploy!

### Docker

Create a `Dockerfile`:
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "run", "start"]
```

Build and run:
```bash
docker build -t digital-memory-frontend .
docker run -p 3000:3000 digital-memory-frontend
```

### Traditional Hosting

Build and deploy the `.next` folder to any Node.js hosting.

## 🔧 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_BASE_URL` | Backend API URL | `http://localhost:8000` |
| `NODE_ENV` | Environment (development/production) | `development` |

## 📚 Additional Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [React Query Documentation](https://tanstack.com/query/latest)

## 🤝 Contributing

Feel free to modify and extend this frontend. Some ideas:
- Add filter UI for source/author/daterange
- Implement dark mode
- Add bookmarks/favorites
- Show embedding similarity visualization
- Add analytics

---

**Ready to start developing! Happy coding! 🚀**
