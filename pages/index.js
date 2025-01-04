export default function Home() {
    return (
        <div style={{
            fontFamily: 'Arial, sans-serif', 
            textAlign: 'center', 
            marginTop: '50px',
        }}>
            <h1 style={{ color: '#0070f3' }}>Welcome to My Simple Next.js Site</h1>
            <p>This is a minimal site created with Next.js and hosted on Vercel.</p>
            <a 
                href="https://nextjs.org" 
                style={{
                    display: 'inline-block',
                    marginTop: '20px',
                    padding: '10px 20px',
                    fontSize: '16px',
                    color: '#ffffff',
                    backgroundColor: '#0070f3',
                    borderRadius: '5px',
                    textDecoration: 'none',
                }}
                target="_blank" 
                rel="noopener noreferrer"
            >
                Learn More About Next.js
            </a>
        </div>
    );
}
