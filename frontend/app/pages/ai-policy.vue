<script setup lang="ts">
    useHead({ title: 'Политика использования контента для AI' });
    useSeoMeta({
        robots: 'noindex, nofollow',
    });
</script>

<template>
    <main class="ai-policy">
        <h1>Политика использования контента для AI</h1>

        <section>
            <h2>О проекте</h2>
            <p>
                «Вечная память героям» — мемориальный проект, посвящённый памяти героев,
                погибших в глобальных и локальных конфликтах. Контент создан в память
                о конкретных людях и их подвигах.
            </p>
        </section>

        <section>
            <h2>Что разрешено</h2>
            <ul>
                <li>Индексация поисковыми системами (Яндекс, Google, Bing)</li>
                <li>Отображение фрагментов в результатах поиска</li>
                <li>Цитирование с указанием источника и ссылки на сайт</li>
            </ul>
        </section>

        <section>
            <h2>Что запрещено</h2>
            <ul>
                <li>Использование контента для обучения моделей искусственного интеллекта</li>
                <li>Включение контента в датасеты для машинного обучения</li>
                <li>Использование контента для тренировки коммерческих и некоммерческих AI-моделей</li>
                <li>Массовое копирование контента (скрейпинг)</li>
                <li>Перераспространение контента без письменного разрешения</li>
            </ul>
        </section>

        <section>
            <h2>Заблокированные боты</h2>
            <p>Доступ полностью закрыт для следующих User-Agent:</p>
            <ul class="ai-policy__bots">
                <li><code>GPTBot</code></li>
                <li><code>OAI-SearchBot</code></li>
                <li><code>ChatGPT-User</code></li>
                <li><code>Google-Extended</code></li>
                <li><code>Google-CloudVertexBot</code></li>
                <li><code>anthropic-ai</code></li>
                <li><code>ClaudeBot</code></li>
                <li><code>Claude-Web</code></li>
                <li><code>cohere-ai</code></li>
                <li><code>PerplexityBot</code></li>
                <li><code>Bytespider</code></li>
                <li><code>Amazonbot</code></li>
                <li><code>Applebot-Extended</code></li>
                <li><code>Meta-ExternalAgent</code></li>
                <li><code>FacebookBot</code></li>
                <li><code>Omgilibot</code></li>
                <li><code>Omgili</code></li>
                <li><code>ImagesiftBot</code></li>
                <li><code>Diffbot</code></li>
                <li><code>YouBot</code></li>
            </ul>
        </section>

        <section>
            <h2>Машиночитаемая политика</h2>
            <p>
                Политика доступна в машиночитаемом формате:
                <NuxtLink to="/.well-known/ai-policy.json">/.well-known/ai-policy.json</NuxtLink>
            </p>
        </section>

        <section>
            <h2>Контакты</h2>
            <p>
                По вопросам использования контента обращайтесь:
                <a href="mailto:admin@emh.su">admin@emh.su</a>
            </p>
        </section>
    </main>
</template>

<style scoped>
    .ai-policy {
        max-width: 48rem;
        margin: 0 auto;
        padding: 3rem 2rem 5rem;
    }

    .ai-policy h1 {
        font-family: var(--font-display);
        font-size: clamp(1.8rem, 4vw, 2.6rem);
        margin-bottom: 2rem;
    }

    .ai-policy section {
        margin-bottom: 2rem;
    }

    .ai-policy h2 {
        font-family: var(--font-display);
        font-size: 1.3rem;
        margin-bottom: 0.75rem;
        padding-bottom: 0.4rem;
        border-bottom: 1px solid var(--emh-line);
    }

    .ai-policy p,
    .ai-policy li {
        font-size: 1rem;
        line-height: 1.7;
        color: var(--emh-ink);
    }

    .ai-policy ul {
        padding-left: 1.5rem;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
    }

    .ai-policy__bots {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
        gap: 0.3rem;
    }

    .ai-policy code {
        font-size: 0.88rem;
        background: rgba(176, 141, 87, 0.08);
        padding: 0.15rem 0.4rem;
        border-radius: 3px;
    }
</style>