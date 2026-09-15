<script setup lang="ts">
    import { defineBreadcrumb, defineOrganization, defineWebPage } from 'nuxt-schema-org/schema';
    import AccountPlusOutlineIcon from '~icons/mdi/account-plus-outline?width=1.25em&height=1.25em';
    import CodeIcon from '~icons/mdi/code?width=1.25em&height=1.25em';
    import EnvelopeOutlineIcon from '~icons/mdi/envelope-outline?width=1.25em&height=1.25em';

    // SEO через nuxt-seo-utils: site.url и site.name читаем из runtimeConfig
    const site = useSiteConfig();
    const canonical = `${site.url}/about`;

    useSeoMeta({
        title: `О проекте`,
        ogTitle: `О проекте — ${site.name}`,
        description:
            '«Вечная память героям» — частный мемориальный проект о героях, погибших в глобальных и локальных конфликтах. ВОВ, Афган, Чечня, Вьетнам, Сирия, Африка, Новороссия, спецоперации.',
        ogDescription:
            '«Вечная память героям» — частный мемориальный проект о героях, погибших в глобальных и локальных конфликтах. ВОВ, Афган, Чечня, Вьетнам, Сирия, Африка, Новороссия, спецоперации.',
        ogType: 'website',
        ogUrl: canonical,
        ogSiteName: site.name,
        ogLocale: 'ru_RU',
        robots: 'index, follow',
    });

    useHead({
        link: [
            { rel: 'canonical', href: canonical },
        ],
    });

    useHead({
        link: [
            {
                rel: 'canonical',
                href: canonical,
            },
        ],
    });

    // ═══════════════════════════════════════════════════════════
    // Schema.org для страницы «О проекте» (nuxt-schema-org)
    // ═══════════════════════════════════════════════════════════
    useSchemaOrg([
        // AboutPage — специальный тип страницы «О проекте»
        defineWebPage({
            '@type': 'AboutPage',
            name: `О проекте — ${site.name}`,
            description: site.description,
            url: canonical,
            inLanguage: 'ru',
            isPartOf: {
                '@type': 'WebSite',
                name: site.name,
                url: site.url,
            },
        }),

        // Автор проекта — частное лицо
        defineOrganization({
            name: 'Thr0TT1e',
            alternateName: 'Автор проекта «Вечная память героям»',
            url: site.url,
            sameAs: ['https://codeberg.org/Thr0TT1e/emh'],
            description:
                'Автор частного мемориального проекта о героях, погибших в глобальных и локальных конфликтах.',
        }),

        // Breadcrumbs для навигации в выдаче
        defineBreadcrumb({
            itemListElement: [
                { name: 'Главная', item: `${site.url}/` },
                { name: 'О проекте', item: canonical },
            ],
        }),
    ]);
</script>

<template>
    <main class="about">
        <!-- ================================================================== -->
        <!-- Шмуцтитул страницы                                                 -->
        <!-- ================================================================== -->
        <header class="about-hero">
            <div class="about-hero__inner reveal" v-reveal>
                <p class="about-hero__eyebrow">
                    <span class="flame" aria-hidden="true" />
                    О проекте
                </p>

                <h1 class="about-hero__title">Вечная память героям</h1>

                <p class="about-hero__lead">
                    «Вечная память героям» — частный мемориальный проект. Он создан,
                    чтобы ни одно имя не исчезло в архивной пыли и ни одна судьба не
                    растворилась в сухой статистике. Каждая запись здесь — человек,
                    который не вернулся с поля боя.
                </p>
            </div>
        </header>

        <!-- ================================================================== -->
        <!-- Основной текст                                                     -->
        <!-- ================================================================== -->
        <section class="about-main">
            <div class="about-main__inner">
                <!-- 1. Миссия -->
                <section class="about-block reveal" v-reveal>
                    <p class="about-block__eyebrow">I. Зачем это нужно</p>

                    <h2>Память, которая не должна прерываться</h2>

                    <p>
                        Войны и конфликты не заканчиваются в день, когда замолкают орудия.
                        Они продолжаются в памяти семей, в письмах, в безымянных могилах и
                        в архивных папках, до которых никто не добрался.
                    </p>

                    <p>
                        Этот проект — попытка вернуть имена из небытия. Мы собираем и
                        сохраняем сведения о героях, погибших в глобальных и локальных
                        конфликтах, чтобы к ним можно было вернуться: назвать по имени,
                        узнать судьбу, передать память дальше.
                    </p>

                    <p>
                        Каждая запись в книге памяти — не строка в реестре, а судьба,
                        собранная по крупицам из открытых источников.
                    </p>
                </section>

                <hr class="rule about-rule" />

                <!-- 2. География и эпохи -->
                <section class="about-block reveal" v-reveal style="transition-delay: 60ms">
                    <p class="about-block__eyebrow">II. Кого мы помним</p>

                    <h2>География и эпохи</h2>

                    <p>
                        Память не имеет границ. Проект объединяет разные эпохи и географии:
                        Великая Отечественная война, Афганистан, Чечня, Вьетнам, Сирия,
                        Африка, Новороссия, спецоперации.
                    </p>

                    <p>
                        Независимо от того, где и когда человек исполнил свой долг, его имя
                        достойно быть сохранённым. Мы не выстраиваем иерархию подвигов:
                        для памяти важен каждый.
                    </p>
                </section>

                <hr class="rule about-rule" />

                <!-- 3. Методология -->
                <section class="about-block reveal" v-reveal style="transition-delay: 120ms">
                    <p class="about-block__eyebrow">III. Как собираются данные</p>

                    <h2>По крупицам</h2>

                    <p>
                        Данные собираются из открытых источников: официальные сводки и
                        публикации, некрологи, материалы СМИ, семейные архивы, памятные
                        знаки, краеведческие исследования, свидетельства сослуживцев и
                        родственников.
                    </p>

                    <p>
                        Каждая запись сопровождается списком источников. Если сведения
                        противоречивы, мы стараемся сохранить обе версии и указать источник
                        каждой, не сглаживая неопределённость.
                    </p>

                    <p>
                        Там, где данных недостаточно, запись остаётся незавершённой — пока
                        не появятся подтверждения.
                    </p>
                </section>

                <hr class="rule about-rule" />

                <!-- 4. Принципы -->
                <section class="about-block reveal" v-reveal style="transition-delay: 180ms">
                    <p class="about-block__eyebrow">IV. Принципы</p>

                    <h2>Каждый важен</h2>

                    <ul class="about-list">
                        <li>
                            <strong>Каждый важен.</strong>
                            Звание, должность и известность не влияют на право быть в книге
                            памяти.
                        </li>

                        <li>
                            <strong>Только факты и источники.</strong>
                            Мы не домысливаем биографии и не заполняем пробелы художественным
                            вымыслом.
                        </li>

                        <li>
                            <strong>Уважение.</strong>
                            Проект не является площадкой для политических оценок: здесь только
                            память и факты.
                        </li>

                        <li>
                            <strong>Открытость.</strong>
                            Исходный код проекта открыт, а данные можно дополнить или
                            исправить.
                        </li>
                    </ul>
                </section>

                <hr class="rule about-rule" />

                <!-- 5. Как помочь -->
                <section class="about-block reveal" v-reveal style="transition-delay: 240ms">
                    <p class="about-block__eyebrow">V. Как помочь</p>

                    <h2>Это общее дело</h2>

                    <p>
                        Если вы знаете человека, который должен быть в книге памяти, или
                        можете уточнить данные, передать фотографии, документы, ссылки на
                        источники — напишите.
                    </p>

                    <ul class="about-links">
                        <li>
                            <NuxtLink to="/submit">
                                <AccountPlusOutlineIcon />
                                Сообщить о герое
                            </NuxtLink>
                        </li>

                        <li>
                            <NuxtLink to="/contacts">
                                <EnvelopeOutlineIcon />
                                Дополнить или исправить сведения
                            </NuxtLink>
                        </li>

                        <li>
                            <a href="https://codeberg.org/Thr0TT1e/emh" target="_blank" rel="noopener noreferrer">
                                <CodeIcon />
                                Открытый код проекта
                            </a>
                        </li>
                    </ul>
                </section>

                <hr class="rule about-rule" />

                <!-- 6. Частный характер проекта -->
                <section class="about-block reveal" v-reveal style="transition-delay: 300ms">
                    <p class="about-block__eyebrow">VI. Важно</p>

                    <h2>Частный мемориальный проект</h2>

                    <p class="about-note">
                        Проект ведётся частным лицом и не является официальной организацией,
                        государственным архивом или СМИ. Это личная инициатива, которая
                        существует благодаря открытым источникам и участию неравнодушных
                        людей.
                    </p>

                    <p>
                        Если вы представляете ведомство, фонд или семейный архив и хотите
                        передать данные для реестра — обязательно укажите это в обращении.
                    </p>
                </section>
            </div>
        </section>
    </main>
</template>

<style scoped>
    .about-hero {
        border-bottom: 1px solid var(--emh-line);
        background: linear-gradient(180deg,
                rgba(255, 255, 255, 0) 0%,
                rgba(163, 22, 33, 0.045) 100%);
    }

    .about-hero__inner,
    .about-main__inner {
        max-width: 72rem;
        margin: 0 auto;
        padding-inline: 2rem;
    }

    .about-hero__inner {
        max-width: 58rem;
        padding-block: 3rem 3.5rem;
        text-align: center;
    }

    .about-hero__eyebrow {
        display: inline-flex;
        align-items: center;
        gap: 0.6rem;
        margin: 0;
        font-size: 0.72rem;
        font-weight: 600;
        letter-spacing: 0.18em;
        text-transform: uppercase;
        color: var(--emh-bronze);
    }

    .about-hero__title {
        margin: 0.4rem 0 1rem;
        font-size: clamp(2.4rem, 5vw, 4rem);
    }

    .about-hero__lead {
        margin: 0 auto;
        max-width: 46rem;
        color: var(--emh-muted);
        font-size: 1.05rem;
    }

    .about-main {
        padding: 3rem 0 5rem;
    }

    .about-main__inner {
        max-width: 56rem;
    }

    .about-block {
        margin: 0;
    }

    .about-block__eyebrow {
        margin: 0 0 0.5rem;
        font-size: 0.72rem;
        font-weight: 600;
        letter-spacing: 0.18em;
        text-transform: uppercase;
        color: var(--emh-bronze);
    }

    .about-block h2 {
        font-size: 1.65rem;
        margin: 0 0 1rem;
    }

    .about-block p {
        margin: 0 0 1rem;
        color: var(--emh-muted);
        line-height: 1.8;
    }

    .about-block p:last-child {
        margin-bottom: 0;
    }

    .about-list {
        margin: 0;
        padding-left: 1.2rem;
        display: grid;
        gap: 0.75rem;
        color: var(--emh-muted);
        line-height: 1.65;
    }

    .about-list li::marker {
        color: var(--emh-bronze);
    }

    .about-links {
        list-style: none;
        margin: 1.2rem 0 0;
        padding: 0;
        display: grid;
        gap: 0.7rem;
    }

    .about-links a {
        display: inline-flex;
        align-items: center;
        gap: 0.6rem;
        font-weight: 600;
        border-bottom: 1px dashed var(--emh-bronze);
        padding-bottom: 0.15rem;
        width: fit-content;
    }

    .about-links a:hover {
        border-bottom-color: var(--emh-crimson);
    }

    .about-links i {
        color: var(--emh-bronze);
    }

    .about-note {
        border-left: 3px solid var(--emh-bronze);
        background: rgba(176, 141, 87, 0.06);
        padding: 1rem 1.2rem;
        color: var(--emh-muted);
    }

    .about-rule {
        margin: 2.8rem 0;
    }

    @media (max-width: 640px) {

        .about-hero__inner,
        .about-main__inner {
            padding-inline: 1rem;
        }

        .about-hero__inner {
            padding-block: 2.2rem 2.6rem;
        }

        .about-main {
            padding: 2rem 0 4rem;
        }

        .about-rule {
            margin: 2rem 0;
        }
    }
</style>