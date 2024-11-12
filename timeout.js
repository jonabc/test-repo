import process from 'process'

process.once('SIGTERM', () => {
    console.log('SIGTERM received')
    process.exit(0)
});

process.once('SIGINT', () => { 
    console.log('SIGINT received')
    process.exit(0)
});

await new Promise(resolve => setTimeout(resolve, 1000 * 60 * 2));
console.log('Timeout finished');